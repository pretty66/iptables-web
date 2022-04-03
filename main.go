/*
 *
 *  * Licensed to the Apache Software Foundation (ASF) under one or more
 *  * contributor license agreements.  See the NOTICE file distributed with
 *  * this work for additional information regarding copyright ownership.
 *  * The ASF licenses this file to You under the Apache License, Version 2.0
 *  * (the "License"); you may not use this file except in compliance with
 *  * the License.  You may obtain a copy of the License at
 *  *
 *  *     http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  * Unless required by applicable law or agreed to in writing, software
 *  * distributed under the License is distributed on an "AS IS" BASIS,
 *  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  * See the License for the specific language governing permissions and
 *  * limitations under the License.
 *
 */

package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/pretty66/iptables-web/pkg/iptables"
	"github.com/pretty66/iptables-web/utils"
	"net/http"
	"os"
	"runtime"
	"time"
)

//go:embed web/index.html
var webIndex []byte

// BuildDate: Binary file compilation time
// BuildVersion: Binary compiled GIT version
var (
	BuildDate    string
	BuildVersion string
)

var (
	username string // IPT_WEB_USERNAME
	password string // IPT_WEB_PASSWORD
	address  string // IPT_WEB_ADDRESS
)

func init() {
	flag.StringVar(&username, "u", "admin", "login username")
	flag.StringVar(&password, "p", "admin", "login password")
	flag.StringVar(&address, "a", ":10001", "http listen address")
	flag.Parse()
	if v := os.Getenv("IPT_WEB_USERNAME"); len(v) > 0 {
		username = v
	}
	if v := os.Getenv("IPT_WEB_PASSWORD"); len(v) > 0 {
		password = v
	}
	if v := os.Getenv("IPT_WEB_ADDRESS"); len(v) > 0 {
		address = v
	}
}

func main() {
	if runtime.GOOS != "linux" {
		panic("Only Linux system is supported")
	}

	ipc, err := iptables.New()
	if err != nil {
		panic(err)
	}

	mux := NewHTTPMux()
	initRoute(mux, ipc)
	server := &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	fmt.Println("listen address:", address)
	fmt.Println("Build Version: ", BuildVersion, "  Date: ", BuildDate)
	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func initRoute(mux *HTTPMux, ipc *iptables.IptablesCMD) {
	mux.Use(auth)
	mux.HandleFunc("/version", func(w http.ResponseWriter, req *http.Request) {
		v, err := ipc.Version()
		utils.Output(w, err, v)
	})

	mux.HandleFunc("/listRule", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		l, err := ipc.ListRule(table, chain)
		utils.Output(w, err, l)
	})

	mux.HandleFunc("/listExec", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		l, err := ipc.ListExec(table, chain)
		utils.Output(w, err, l)
	})

	mux.HandleFunc("/flushRule", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		err := ipc.FlushRule(table, chain)
		utils.Output(w, err, nil)
	})

	mux.HandleFunc("/deleteRule", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		id := req.FormValue("id")
		err := ipc.DeleteRule(table, chain, id)
		utils.Output(w, err, nil)
	})

	mux.HandleFunc("/flushMetrics", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		id := req.FormValue("id")
		err := ipc.FlushMetrics(table, chain, id)
		utils.Output(w, err, nil)
	})

	mux.HandleFunc("/getRuleInfo", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		id := req.FormValue("id")
		info, err := ipc.GetRuleInfo(table, chain, id)
		utils.Output(w, err, info)
	})

	mux.HandleFunc("/flushEmptyCustomChain", func(w http.ResponseWriter, req *http.Request) {
		err := ipc.FlushEmptyCustomChain()
		utils.Output(w, err, nil)
	})

	mux.HandleFunc("/export", func(w http.ResponseWriter, req *http.Request) {
		table := req.FormValue("table")
		chain := req.FormValue("chain")
		rule, err := ipc.Export(table, chain)
		utils.Output(w, err, rule)
	})

	mux.HandleFunc("/import", func(w http.ResponseWriter, req *http.Request) {
		rule := req.FormValue("rule")
		err := ipc.Import(rule)
		utils.Output(w, err, nil)
	})

	mux.HandleFunc("/exec", func(w http.ResponseWriter, req *http.Request) {
		args := req.FormValue("args")
		if len(args) == 0 {
			utils.Output(w, nil, nil)
			return
		}
		s := utils.SplitAndTrimSpace(args, " ")
		str, err := ipc.Exec(s...)
		utils.Output(w, err, str)
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html;charset-utf8;")
		_, _ = w.Write(webIndex)
	})
}

func auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok && user == username && pass == password {
			handler.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

type Middleware func(http.Handler) http.Handler

type HTTPMux struct {
	*http.ServeMux
	middlewares []Middleware
}

func NewHTTPMux() *HTTPMux {
	return &HTTPMux{
		ServeMux: http.NewServeMux(),
	}
}

func (m *HTTPMux) Use(middlewares ...Middleware) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *HTTPMux) Handle(pattern string, handler http.Handler) {
	handler = applyMiddlewares(handler, m.middlewares...)
	m.ServeMux.Handle(pattern, handler)
}

func (m *HTTPMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	newHandler := applyMiddlewares(handler, m.middlewares...)
	m.ServeMux.Handle(pattern, newHandler)
}

func applyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
