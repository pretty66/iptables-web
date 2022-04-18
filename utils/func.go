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
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

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

func FloatToString(Num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(Num, 'f', 2, 64)
}

func MD5Bytes(s []byte) string {
	ret := md5.Sum(s)
	return hex.EncodeToString(ret[:])
}

func MD5(s string) string {
	return MD5Bytes([]byte(s))
}

func MD5File(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return MD5Bytes(data), nil
}

func ToInterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil
	}
	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GoWithRecover wraps a `go func()` with recover()
func GoWithRecover(handler func(), recoverHandler func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("%s goroutine panic: %v\n%s\n", time.Now().Format("2006-01-02 15:04:05"), r, string(debug.Stack()))
				if recoverHandler != nil {
					go func() {
						defer func() {
							if p := recover(); p != nil {
								log.Println("recover goroutine panic:%v\n%s\n", p, string(debug.Stack()))
							}
						}()
						recoverHandler(r)
					}()
				}
			}
		}()
		handler()
	}()
}

func TimeMs() int64 {
	return time.Now().UnixNano() / 1e6
}

// JsonEncode
func JsonEncode(param interface{}) []byte {
	b, err := json.Marshal(param)
	if err != nil {
		pc, f, l, _ := runtime.Caller(1)
		fc := runtime.FuncForPC(pc)
		log.Printf("json_encode err: file:%s, line:%s, function name:%s, err:%s", f, l, fc.Name(), err.Error())
		return []byte{}
	}
	return b
}

func StructToMap(obj interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)
	data := make(map[string]interface{})
	for i := 0; i < obj1.NumField(); i++ {
		data[obj1.Field(i).Name] = obj2.Field(i).Interface()
	}
	return data
}

// Convert json string to map
func JsonToMap(jsonStr string) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func GetInterfaceString(param interface{}) string {
	switch param.(type) {
	case string:
		return param.(string)
	case int:
		return strconv.Itoa(param.(int))
	case float64:
		return strconv.Itoa(int(param.(float64)))
	}
	return ""
}

func NewMd5(str ...string) string {
	h := md5.New()
	for _, v := range str {
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func InArray(in string, array []string) bool {
	for k := range array {
		if in == array[k] {
			return true
		}
	}
	return false
}

func MapDeepCopy(value map[string]string) map[string]string {
	newMap := make(map[string]string)
	if value == nil {
		return newMap
	}
	for k, v := range value {
		newMap[k] = v
	}

	return newMap
}

func RemoveDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func DumpCertAndPrivateKey(cert *tls.Certificate) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Leaf.Raw,
	}
	log.Println(string(pem.EncodeToMemory(block)))
	b, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		log.Println("x509.MarshalPKCS8PrivateKey", err)
		return
	}
	log.Println(string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})))
}

func DumpUnit(l int) string {
	if l < 1024 {
		return fmt.Sprintf("%dB", l)
	} else if l < 1048576 {
		return fmt.Sprintf("%dKB", l/1024)
	} else {
		return fmt.Sprintf("%dMb", l/1048576)
	}
}

// GenerateUUID generates an uuid
// https://tools.ietf.org/html/rfc4122
// crypto.rand use getrandom(2) or /dev/urandom
// It is maybe occur an error due to system error
// panic if an error occurred
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		panic("generate an uuid failed, error: " + err.Error())
	}
	// see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
