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
	AnObject
}

type SwitchItem struct {
	AnItem
}

func (s *Switch) on() string {
	Log.Infof("Switch %s set to ON", s.ID())

	old := s.AnObject.SetState(ON)

	s.notifyListeners(old, ON)

	return old
}

func (s *Switch) off() string {
	Log.Infof("Switch %s set to OFF", s.ID())

	old := s.AnObject.SetState(OFF)

	s.notifyListeners(old, OFF)

	return old
}

func (s *Switch) SetState(new string) string {
	var old string
	switch new {
	case "on", "ON", "1":
		old = s.on()
	default:
		old = s.off()
	}

	return old
}

func (si *SwitchItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(si)
}

func newSwitch(id string, label string, device interface{}) *Switch {
	s := &Switch{
		AnObject: AnObject{
			id:     id,
			label:  label,
			device: device,
			items:  make(map[string]Item),
			state:  OFF,
		},
	}

	s.items[ItemID] = &SwitchItem{
		AnItem: AnItem{
			object: s,
			kind:   "switch",
			img:    "switch",
		},
	}

	return s
}

// RegisterSwitch register a new Switch. A switch has two states ON or OFF.
// It will emit a message on the bus when changing state from ON to OFF and
// vise versa.
func RegisterSwitch(id string, label string, device interface{}) *Switch {
	s := newSwitch(id, label, device)
	RegisterObject(s)
	return s
}
