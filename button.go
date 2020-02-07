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

type Button struct {
	AnItem
}

func (b *Button) push() (string, bool) {
	old, updated := b.AnItem.SetValue(ON)
	b.notifyListeners(old, ON)

	return old, updated
}

func (b *Button) SetValue(new string) (string, bool) {
	return b.push()
}

func NewButton(id string, label string) *Button {
	s := &Button{
		AnItem: AnItem{
			id:    id,
			label: label,
			kind:  "button",
			img:   "switch",
		},
	}

	registry.Add(s)

	return s
}
