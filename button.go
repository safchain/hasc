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

type Button struct {
	AnObject
}

type ButtonItem struct {
	AnItem
}

func (b *Button) push() string {
	Log.Infof("Button %s pushed", b.ID())

	old := b.AnObject.SetState(ON)

	b.notifyListeners("", ON)

	return old
}

func (b *Button) SetState(new string) string {
	return b.push()
}

func (bi *ButtonItem) HTML() template.HTML {
	data := struct {
		ID       string
		ObjectID string
		Label    string
		Img      string
	}{
		ID:       bi.object.ID() + "_" + bi.ID(),
		ObjectID: bi.object.ID(),
		Label:    bi.object.Label(),
		Img:      fmt.Sprintf("statics/img/%s.png", bi.img),
	}

	return itemTemplate("statics/items/button.html", data)
}

func (bi *ButtonItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(bi)
}

func newButton(id string, label string, device interface{}) *Button {
	s := &Button{
		AnObject: AnObject{
			id:     id,
			label:  label,
			device: device,
			items:  make(map[string]Item),
		},
	}

	s.items[ItemID] = &ButtonItem{
		AnItem: AnItem{
			object: s,
			img:    "switch",
		},
	}

	return s
}

// RegisterButton registers a new button using the given ID, label and device
// The object will send ON state to the bus.
func RegisterButton(id string, label string, device interface{}) *Button {
	s := newButton(id, label, device)
	RegisterObject(s)
	return s
}
