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

package group

import (
	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type GroupItem struct {
	item.AnItem

	Items []item.Item
}

func (g *GroupItem) refresh() {
	off := g.GetValue() == item.OFF || g.GetValue() == ""

	new := item.OFF
	for _, it := range g.Items {
		os := it.GetValue()
		if os != item.OFF && os != "" {
			new = item.ON

			if off {
				g.AnItem.SetValue(item.ON)
				break
			}
		}
	}

	if new == item.OFF && !off {
		g.AnItem.SetValue(item.OFF)
	}
}

func (g *GroupItem) OnValueChange(it item.Item, old string, new string) {
	g.refresh()
}

// Add adds an Object to the group.
func (g *GroupItem) Add(it item.Item) {
	g.Items = append(g.Items, it)

	it.AddListener(g)

	g.refresh()
}

func NewGroupItem(id string, label string) *GroupItem {
	g := &GroupItem{
		AnItem: item.AnItem{
			ID:    id,
			Label: label,
			Type:  "state",
			Img:   "group",
		},
	}
	g.SetValue(item.OFF)

	server.Registry.Add(g, "")

	return g
}
