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
	"fmt"
	"html/template"
)

type Switch struct {
	AnObject
}

type SwitchItem struct {
	AnItem
}

func (s *Switch) on() {
	Log.Infof("Switch %s set to ON", s.ID())

	s.Lock()
	old := s.state
	s.state = ON
	s.Unlock()

	s.notifyListeners(old, ON)
}

func (s *Switch) off() {
	Log.Infof("Switch %s set to OFF", s.ID())

	s.Lock()
	old := s.state
	s.state = OFF
	s.Unlock()

	s.notifyListeners(old, OFF)
}

func (s *Switch) SetState(new string) {
	switch new {
	case "on", "ON", "1":
		s.on()
	default:
		s.off()
	}
}

func (si *SwitchItem) HTML() template.HTML {
	data := struct {
		ID       string
		ObjectID string
		Label    string
		Img      string
	}{
		ID:       si.object.ID() + "_" + si.ID(),
		ObjectID: si.object.ID(),
		Label:    si.object.Label(),
		Img:      fmt.Sprintf("statics/img/%s.png", si.img),
	}

	return itemTemplate("statics/items/switch.html", data)
}

func (si *SwitchItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(si)
}

func newSwitch(id string, label string, address1 int, address2 int, receiver int) *Switch {
	s := &Switch{
		AnObject: AnObject{
			id:       id,
			label:    label,
			address1: address1,
			address2: address2,
			receiver: receiver,
			items:    make(map[string]Item),
			state:    OFF,
		},
	}

	s.items[ItemID] = &SwitchItem{
		AnItem: AnItem{
			object: s,
			img:    "switch",
		},
	}

	return s
}

// RegisterSwitch register a new Switch. A switch has two states ON or OFF.
// It will emit a message on the bus when changing state from ON to OFF and
// vise versa.
func RegisterSwitch(id string, label string, address1 int, address2 int, receiver int) *Switch {
	s := newSwitch(id, label, address1, address2, receiver)
	RegisterObject(s)
	return s
}
