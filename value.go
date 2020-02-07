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

type Value struct {
	AnItem
}

func (v *Value) SetValue(new string) (string, bool) {
	old, updated := v.AnItem.SetValue(new)
	v.notifyListeners(old, new)

	return old, updated
}

func NewValue(id string, label string, units ...string) *Value {
	var unit string
	if len(units) > 0 {
		unit = units[0]
	}

	v := &Value{
		AnItem: AnItem{
			id:    id,
			label: label,
			kind:  "value",
			img:   "chart",
			unit:  unit,
		},
	}

	registry.Add(v)

	return v
}
