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

type Group struct {
	AnItem
	items []Item
}

func (g *Group) on() {
	old, updated := g.AnItem.SetValue(ON)
	if updated {
		g.notifyListeners(old, ON)
	}
}

func (g *Group) off() {
	old, updated := g.AnItem.SetValue(OFF)
	if updated {
		g.notifyListeners(old, OFF)
	}
}

func (g *Group) refresh() {
	g.RLock()
	off := g.value == OFF || g.value == ""
	g.RUnlock()

	new := OFF
	for _, item := range g.items {
		os := item.Value()
		if os != OFF && os != "" {
			new = ON

			if off {
				g.on()
				break
			}
		}
	}

	if new == OFF && !off {
		g.off()
	}
}

func (g *Group) OnValueChange(item Item, old string, new string) {
	g.refresh()
}

// Add adds an Object to the group.
func (g *Group) Add(item Item) {
	g.items = append(g.items, item)

	item.AddListener(g)

	g.refresh()
}

func (g *Group) Items() []Item {
	return g.items
}

func NewGroup(id string, label string) *Group {
	g := &Group{
		AnItem: AnItem{
			id:    id,
			label: label,
			kind:  "state",
			img:   "group",
			value: OFF,
		},
	}

	registry.Add(g)

	return g
}
