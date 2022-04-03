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

package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

func SplitAndTrimSpace(s, sep string) []string {
	res := strings.Split(s, sep)
	for k := range res {
		res[k] = strings.TrimSpace(res[k])
	}
	return res
}

func JSONEncoding(data interface{}) []byte {
	b, _ := json.Marshal(data)
	return b
}

func Output(w http.ResponseWriter, err error, data interface{}) {
	var code int
	msg := "OK"
	if err != nil {
		code = 1
		msg = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
	out := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	}
	_, _ = w.Write(JSONEncoding(out))
}
