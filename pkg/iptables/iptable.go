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
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pretty66/iptables-web/utils"
)

type Protocol byte

const (
	ProtocolIPv4 Protocol = iota
	ProtocolIPv6
)

type IptablesV4CMD struct {
	binary        string
	saveBinary    string
	restoreBinary string
	protocol      Protocol
	exec          exec.Cmd
}

type option func(*IptablesV4CMD)

func NewIPV4(opt ...option) (*IptablesV4CMD, error) {
	ipc := &IptablesV4CMD{}
	for _, fn := range opt {
		fn(ipc)
	}
	if ipc.protocol == ProtocolIPv6 {
		return nil, errors.New("IPv6 is not supported temporarily")
	}
	if len(ipc.binary) == 0 {
		ipc.binary = "iptables"
	}
	if len(ipc.saveBinary) == 0 {
		ipc.saveBinary = "iptables-save"
	}
	if len(ipc.restoreBinary) == 0 {
		ipc.restoreBinary = "iptables-restore"
	}
	return ipc, nil
}

func WithProtocol(protocol Protocol) option {
	return func(ic *IptablesV4CMD) {
		ic.protocol = protocol
	}
}

func WithBinary(cmd string) option {
	return func(ic *IptablesV4CMD) {
		ic.binary = cmd
	}
}

func WithSaveBinary(cmd string) option {
	return func(ic *IptablesV4CMD) {
		ic.saveBinary = cmd
	}
}

func WithRestoreBinary(cmd string) option {
	return func(ic *IptablesV4CMD) {
		ic.restoreBinary = cmd
	}
}

func (i *IptablesV4CMD) Version() (string, error) {
	return i.iptables("--version")
}

func (i *IptablesV4CMD) ListRule(table, chain string) (map[string][]TableList, error) {
	if len(table) == 0 {
		table = "filter"
	}
	var str string
	var err error
	if len(chain) == 0 {
		str, err = i.iptables("-t", table, "-nvL", "--line-numbers")
	} else {
		str, err = i.iptables("-t", table, "-L", chain, "-nv", "--line-numbers")
	}

	if err != nil {
		return nil, err
	}

	tl := map[string][]TableList{}
	tl["system"] = make([]TableList, 0)
	tl["custom"] = make([]TableList, 0)

	chains := utils.SplitAndTrimSpace(str, "\n\n")
	for k := range chains {
		column := []Column{}
		chainList := utils.SplitAndTrimSpace(chains[k], "\n")
		if len(chainList) == 0 {
			continue
		}
		if len(chainList) > 2 {
			column, err = parseColumn(chainList[2:])
			if err != nil {
				log.Println(err)
				continue
			}
		}

		stitle, err := parseSystemTitle(chainList[0])
		if err == nil {
			tl["system"] = append(tl["system"], SystemTable{
				SystemTitle: stitle,
				Column:      column,
			})
		} else {
			ctitle, err := parseCustomTitle(chainList[0])
			if err != nil {
				log.Println(err)
				continue
			}
			tl["custom"] = append(tl["custom"], CustomTable{
				CustomTitle: ctitle,
				Column:      column,
			})
		}
	}
	return tl, nil
}

func (i *IptablesV4CMD) FlushRule(table, chain string) error {
	var err error
	if len(table) == 0 && len(chain) == 0 {
		_, err = i.iptables("-t", "raw", "-F")
		_, err = i.iptables("-t", "mangle", "-F")
		_, err = i.iptables("-t", "nat", "-F")
		_, err = i.iptables("-t", "filter", "-F")
		return err
	}

	if len(table) == 0 {
		table = "filter"
	}
	if len(chain) == 0 {
		_, err = i.iptables("-t", table, "-F")
	} else {
		_, err = i.iptables("-t", table, "-F", chain)
	}

	return err
}

func (i *IptablesV4CMD) FlushMetrics(table, chain, id string) error {
	var err error
	if len(id) > 0 {
		if len(table) == 0 || len(chain) == 0 {
			return fmt.Errorf("FlushMetrics args error. table:%s chain:%s id:%s", table, chain, id)
		}
		_, err = i.iptables("-t", table, "-Z", chain, id)
		return err
	}

	if len(table) == 0 && len(chain) == 0 {
		_, err = i.iptables("-t", "raw", "-Z")
		_, err = i.iptables("-t", "mangle", "-Z")
		_, err = i.iptables("-t", "nat", "-Z")
		_, err = i.iptables("-t", "filter", "-Z")
		return err
	}

	if len(table) == 0 {
		table = "filter"
	}
	if len(chain) == 0 {
		_, err = i.iptables("-t", table, "-Z")
	} else {
		_, err = i.iptables("-t", table, "-Z", chain)
	}

	return err
}

func (i *IptablesV4CMD) DeleteRule(table, chain, id string) error {
	if len(table) == 0 || len(chain) == 0 || len(id) == 0 {
		return fmt.Errorf("DeleteRule args error. table:%s chain:%s id:%s", table, chain, id)
	}
	_, err := i.iptables("-t", table, "-D", chain, id)
	return err
}

func (i *IptablesV4CMD) ListExec(table, chain string) (string, error) {
	var str string
	var err error
	if len(chain) == 0 {
		str, err = i.iptablesSave("-t", table)
	} else {
		// chain不用去除空格，显示引用命令
		str, err = i.iptablesSave("-t", table, "|", "grep", chain)
	}
	if err != nil {
		log.Println("ListExec:", err)
	}
	return str, err
}

func (i *IptablesV4CMD) Exec(param ...string) (string, error) {
	var args []string
	for k := range param {
		param[k] = strings.TrimSpace(param[k])
		if len(param[k]) == 0 {
			continue
		}
		args = append(args, param[k])
	}
	return i.iptables(args...)
}

func (i *IptablesV4CMD) GetRuleInfo(table, chain, id string) (string, error) {
	if len(table) == 0 || len(chain) == 0 || len(id) == 0 {
		return "", fmt.Errorf("GetRuleInfo args error. table:%s chain:%s id:%s", table, chain, id)
	}
	// s, err := i.iptablesSave(fmt.Sprintf("-t %s | grep %s", table, " "+chain+" "))
	s, err := i.iptablesSave(fmt.Sprintf("-t %s | grep ' %s '", table, chain))
	if err != nil {
		return "", err
	}
	list := utils.SplitAndTrimSpace(s, "\n")
	idint, _ := strconv.Atoi(id)
	if len(list) < idint {
		return "", fmt.Errorf("GetRuleInfo rule not found. table:%s chain:%s id:%s", table, chain, id)
	}
	return list[idint-1], nil
}

func (i *IptablesV4CMD) FlushEmptyCustomChain() error {
	_, err := i.iptables("-t", "raw", "-X")
	_, err = i.iptables("-t", "mangle", "-X")
	_, err = i.iptables("-t", "nat", "-X")
	_, err = i.iptables("-t", "filter", "-X")
	return err
}

func (i *IptablesV4CMD) Export(table, chain string) (string, error) {
	var args []string
	if len(table) > 0 {
		args = append(args, table)
	}
	if len(chain) > 0 {
		args = append(args, chain)
	}
	return i.iptablesSave(args...)
}

func (i *IptablesV4CMD) Import(rule string) error {
	if len(rule) == 0 {
		return nil
	}
	fileName := "/tmp/iptable.rule"
	err := os.WriteFile(fileName, []byte(rule), fs.ModePerm)
	if err != nil {
		return fmt.Errorf("Import rule error. err:%v", err)
	}
	defer os.Remove(fileName)
	_, err = i.iptablesRestore(fileName)
	return err
}

func (i *IptablesV4CMD) iptables(args ...string) (string, error) {
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(i.binary, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("exec: [%s %s] err: %v", i.binary, strings.Join(args, " "), errBuf.String())
	}
	return strings.TrimSpace(outBuf.String()), nil
}

func (i *IptablesV4CMD) iptablesSave(args ...string) (string, error) {
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", i.saveBinary, strings.Join(args, " ")))
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("exec: [sh -c %s] err: %s", fmt.Sprintf("%s %s", i.saveBinary, strings.Join(args, " ")), errBuf.String())
	}
	return strings.TrimSpace(outBuf.String()), nil
}

func (i *IptablesV4CMD) iptablesRestore(fileName string) (string, error) {
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s < %s", i.restoreBinary, fileName))
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("exec: [%s < %s] err: %s", i.restoreBinary, fileName, errBuf.String())
	}
	return strings.TrimSpace(outBuf.String()), nil
}
