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

import "html/template"

type Value struct {
	AnObject
}

type ValueItem struct {
	AnItem
}

func (v *Value) SetState(new string) {
	Log.Infof("Sensor %s set to %s", v.ID(), new)

	v.Lock()
	old := v.state
	v.state = new
	v.Unlock()

	v.notifyListeners(old, new)
}

func (vi *ValueItem) HTML() template.HTML {
	vi.object.RLock()
	defer vi.object.RUnlock()

	return valueTemplate(vi, "", "", vi.img)
}

func (vi *ValueItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(vi)
}

func newValue(id string, label string) *Value {
	v := &Value{
		AnObject: AnObject{
			id:    id,
			label: label,
			items: make(map[string]Item),
		},
	}

	v.items[ItemID] = &ValueItem{
		AnItem: AnItem{
			object: v,
			img:    "chart",
		},
	}

	return v
}

// RegisterValue registers a simple value object.
func RegisterValue(id string, label string) *Value {
	v := newValue(id, label)
	RegisterObject(v)
	return v
}
