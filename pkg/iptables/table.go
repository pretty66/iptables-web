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

package iptables

import (
	"fmt"
	"regexp"

	"github.com/pretty66/iptables-web/utils"
)

type SystemTitle struct {
	Chain   string `json:"chain"`   // PREROUTING、INPUT、OUTPUT、POSTROUTING
	Policy  string `json:"policy"`  // 默认策略：ACCEPT、DROP
	Packets string `json:"packets"` // 包数量
	Bytes   string `json:"bytes"`   // 字节数
}

type CustomTitle struct {
	Chain      string `json:"chain"`      // PREROUTING、INPUT、OUTPUT、POSTROUTING
	References string `json:"references"` // 引用数量
}

type Column struct {
	Num         string `json:"num"`
	Pkts        string `json:"pkts"`
	Bytes       string `json:"bytes"`
	Target      string `json:"target"`
	Prot        string `json:"prot"`
	Opt         string `json:"opt"`
	In          string `json:"in"`
	Out         string `json:"out"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Action      string `json:"action"`
}

type SystemTable struct {
	SystemTitle `json:"title"`
	Column      []Column `json:"list"`
}

func (st SystemTable) String() string {
	return string(utils.JSONEncoding(st))
}

type CustomTable struct {
	CustomTitle `json:"title"`
	Column      []Column `json:"list"`
}

func (st CustomTable) String() string {
	return string(utils.JSONEncoding(st))
}

type TableList interface {
	String() string
}

var (
	systemTitleRegex *regexp.Regexp
	customTitleRegex *regexp.Regexp
	columnRegex      *regexp.Regexp
)

func init() {
	var err error
	systemTitleRegex, err = regexp.Compile(`Chain (.+) \(policy (.+) (.+) packets, (.+) bytes\)`)
	if err != nil {
		panic(err)
	}
	customTitleRegex, err = regexp.Compile(`Chain (.+) \((\d+) references\)`)
	if err != nil {
		panic(err)
	}
	columnRegex, err = regexp.Compile(`(\d+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+([0-9\.\/]+)\s*(.*)`)
	if err != nil {
		panic(err)
	}
}

func parseSystemTitle(ts string) (out SystemTitle, err error) {
	res := systemTitleRegex.FindStringSubmatch(ts)
	if len(res) != 5 {
		err = fmt.Errorf("parse system table title error:%d => %v", len(res), res)
		return
	}
	out.Chain = res[1]
	out.Policy = res[2]
	out.Packets = res[3]
	out.Bytes = res[4]
	return
}

func parseCustomTitle(ts string) (out CustomTitle, err error) {
	res := customTitleRegex.FindStringSubmatch(ts)
	if len(res) != 3 {
		err = fmt.Errorf("parse custom table title error:%d => %v", len(res), res)
		return
	}
	out.Chain = res[1]
	out.References = res[2]
	return
}

func parseColumn(cs []string) ([]Column, error) {
	out := []Column{}
	for k := range cs {
		if len(cs[k]) == 0 {
			continue
		}
		rule := columnRegex.FindStringSubmatch(cs[k])
		if len(rule) < 12 {
			return nil, fmt.Errorf("parse column error:%d => %v, str:%s", len(rule), rule, cs[k])
		}
		rule = rule[1:]
		c := Column{
			Num:         rule[0],
			Pkts:        rule[1],
			Bytes:       rule[2],
			Target:      rule[3],
			Prot:        rule[4],
			Opt:         rule[5],
			In:          rule[6],
			Out:         rule[7],
			Source:      rule[8],
			Destination: rule[9],
			Action:      rule[10],
		}
		out = append(out, c)
	}
	return out, nil
}
