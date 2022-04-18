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
	"errors"
	"os/exec"
)

type IptablesV6CMD struct {
	binary        string
	saveBinary    string
	restoreBinary string
	protocol      Protocol
	exec          exec.Cmd
}

func (i *IptablesV6CMD) NewIPV6() (*IptablesV6CMD, error) {
	return nil, errors.New("waiting to be realized")
}

func (i *IptablesV6CMD) Version() (string, error) {
	return "", nil
}

func (i *IptablesV6CMD) ListRule(table, chain string) (map[string][]TableList, error) {
	return nil, nil
}

func (i *IptablesV6CMD) FlushRule(table, chain string) error {
	return nil
}

func (i *IptablesV6CMD) FlushMetrics(table, chain, id string) error {
	return nil
}

func (i *IptablesV6CMD) DeleteRule(table, chain, id string) error {
	return nil
}

func (i *IptablesV6CMD) ListExec(table, chain string) (string, error) {
	return "", nil
}

func (i *IptablesV6CMD) Exec(param ...string) (string, error) {
	return "", nil
}

func (i *IptablesV6CMD) GetRuleInfo(table, chain, id string) (string, error) {
	return "", nil
}

func (i *IptablesV6CMD) FlushEmptyCustomChain() error {
	return nil
}

func (i *IptablesV6CMD) Export(table, chain string) (string, error) {
	return "", nil
}

func (i *IptablesV6CMD) Import(rule string) error {
	return nil
}
