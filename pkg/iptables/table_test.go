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
	"strings"
	"testing"
)

const ruleList = `Chain PREROUTING (policy ACCEPT 2188 packets, 652K bytes)
num   pkts bytes target     prot opt in     out     source               destination
1       24  1604 DOCKER     all  --  *      *       0.0.0.0/0            0.0.0.0/0            ADDRTYPE match dst-type LOCAL

Chain INPUT (policy ACCEPT 1163 packets, 590K bytes)
num   pkts bytes target     prot opt in     out     source               destination

Chain OUTPUT (policy ACCEPT 11032 packets, 669K bytes)
num   pkts bytes target     prot opt in     out     source               destination
1       31  1860 DOCKER     all  --  *      *       0.0.0.0/0           !127.0.0.0/8          ADDRTYPE match dst-type LOCAL

Chain POSTROUTING (policy ACCEPT 11095 packets, 673K bytes)
num   pkts bytes target     prot opt in     out     source               destination
1        0     0 MASQUERADE  all  --  *      !docker0  172.17.0.0/16        0.0.0.0/0
2      126  7560 MASQUERADE  all  --  *      !br-152b31f15eed  172.19.0.0/16        0.0.0.0/0
3        0     0 MASQUERADE  tcp  --  *      *       172.19.0.8           172.19.0.8           tcp dpt:8083
4        0     0 MASQUERADE  tcp  --  *      *       172.19.0.8           172.19.0.8           tcp dpt:8082
5        0     0 MASQUERADE  tcp  --  *      *       172.19.0.8           172.19.0.8           tcp dpt:8081
6        0     0 MASQUERADE  tcp  --  *      *       172.19.0.9           172.19.0.9           tcp dpt:3306

Chain DOCKER (2 references)
num   pkts bytes target     prot opt in     out     source               destination
1        0     0 RETURN     all  --  docker0 *       0.0.0.0/0            0.0.0.0/0
2        0     0 RETURN     all  --  br-152b31f15eed *       0.0.0.0/0            0.0.0.0/0
3        0     0 DNAT       tcp  --  !br-152b31f15eed *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8083 to:172.19.0.8:8083
4        0     0 DNAT       tcp  --  !br-152b31f15eed *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8082 to:172.19.0.8:8082
5        0     0 DNAT       tcp  --  !br-152b31f15eed *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8081 to:172.19.0.8:8081
6        0     0 DNAT       tcp  --  !br-152b31f15eed *       0.0.0.0/0            0.0.0.0/0            tcp dpt:3306 to:172.19.0.9:3306
`

func TestParse(t *testing.T) {
	chainList := strings.Split(ruleList, "\n\n")

	for k := range chainList {
		chainInfo := strings.Split(chainList[k], "\n")
		if len(chainInfo) == 0 {
			continue
		}
		title := chainInfo[0]
		column := chainInfo[1]
		t.Log(title)
		t.Log(column, len(chainInfo)-2)
	}
}

func TestParseSystemTitle(t *testing.T) {
	list := []string{
		`Chain POSTROUTING (policy ACCEPT 11095 packets, 673K bytes)`,
		`Chain INPUT (policy ACCEPT 1163 packets, 590K bytes)`,
		`Chain OUTPUT (policy ACCEPT 11032 packets, 669K bytes)`,
		`Chain POSTROUTING (policy ACCEPT 11095 packets, 673K bytes)`,
	}
	for k := range list {
		title, err := parseSystemTitle(list[k])
		if err != nil {
			t.Fatal(err)
		}
		t.Log(title)
	}
}

func TestParseCustomTitle(t *testing.T) {
	list := []string{
		`Chain DOCKER (2 references)`,
	}
	for k := range list {
		title, err := parseCustomTitle(list[k])
		if err != nil {
			t.Fatal(err)
		}
		t.Log(title)
	}
}

func TestParseColumn(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		line := `1        0     0 MASQUERADE  all  --  *      !docker0  172.17.0.0/16        0.0.0.0/0   aaaaa aaa aaa`
		res := columnRegex.FindStringSubmatch(line)
		t.Log(res)
		if len(res) != 12 {
			t.Error("match column error")
		}
	})

	t.Run("list", func(t *testing.T) {
		list := []string{
			`2      126  7560 MASQUERADE  all  --  *      !br-152b31f15eed  172.19.0.0/16        0.0.0.0/0  `,
			`2      126  7560 MASQUERADE  all  --  *      !br-152b31f15eed  172.19.0.0/16        0.0.0.0/0  `,
			`3        0     0 MASQUERADE  tcp  --  *      *       172.19.0.8           172.19.0.8           tcp dpt:8083`,
		}
		csl, err := parseColumn(list)
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("%#v", csl)
	})

	//t.Run("iptables", func(t *testing.T) {
	//	ipc, _ := New()
	//	s := ipc.iptables("-t", "nat", "-L", "POSTROUTING", "-nv", "--line-numbers")
	//	list := strings.Split(s, "\n")
	//
	//	csl, err := parseColumn(list[2:])
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//	t.Log(csl)
	//})
}
