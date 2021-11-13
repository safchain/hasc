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

type Switch struct {
	AnItem
}

func (s *Switch) on() (string, bool) {
	old, updated := s.AnItem.SetValue(ON)
	if updated {
		s.notifyListeners(old, ON)
	}

	return old, updated
}

func (s *Switch) off() (string, bool) {
	old, updated := s.AnItem.SetValue(OFF)
	if updated {
		s.notifyListeners(old, OFF)
	}

	return old, updated
}

func (s *Switch) SetValue(new string) (string, bool) {
	var old string
	var updated bool

	switch new {
	case "on", "ON", "1":
		old, updated = s.on()
	default:
		old, updated = s.off()
	}

	return old, updated
}

func NewSwitch(id string, label string, disabled bool) *Switch {
	kind := "switch"
	if disabled {
		kind = "state"
	}

	s := &Switch{
		AnItem: AnItem{
			id:    id,
			label: label,
			kind:  kind,
			img:   "switch",
			value: OFF,
		},
	}

	registry.Add(s)

	return s
}
