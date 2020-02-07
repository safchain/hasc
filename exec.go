/*
 * Copyright (C) 2017 Sylvain Afchain
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package hasc

import (
	"os/exec"
	"strings"
)

type Exec struct {
	cmdOn  []string
	cmdOff []string
	item   Item
}

func (e *Exec) OnStateChange(id string, old string, new string) {
	var cmd *exec.Cmd
	var arg0 string
	var args []string

	switch new {
	case "on", "ON", "1":
		arg0 = e.cmdOn[0]
		args = e.cmdOn[1:]
	default:
		arg0 = e.cmdOff[0]
		args = e.cmdOff[1:]
	}

	cmd = exec.Command(arg0, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		Log.Errorf("Exec command `%s %s` error: %s", arg0, strings.Join(args, " "), err)
		return
	}

	if e.item != nil {
		e.item.SetValue(string(output))
	}
}

// NewExec returns a new Exec object. The given cmdOn parameter will be used while the
// object state will be set to ON and the cmdoff for the OFF state of the object.
// The extra obj parameter can be used to receive the output of the execution.
func NewExec(cmdOn []string, cmdOff []string, items ...Item) *Exec {
	Log.Infof("New Exec on: %s, off: %s", cmdOn, cmdOff)

	e := &Exec{
		cmdOn:  cmdOn,
		cmdOff: cmdOff,
	}

	if len(items) > 0 {
		e.item = items[0]
	}

	return e
}
