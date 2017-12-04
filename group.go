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
	"encoding/json"
	"fmt"
	"html/template"
)

type Group struct {
	AnObject
	objects []Object
}

type GroupItem struct {
	AnItem
}

func (g *Group) on() {
	Log.Infof("Group %s set to ON", g.ID())

	g.Lock()
	old := g.state
	g.state = ON
	g.Unlock()

	g.notifyListeners(old, ON)
}

func (g *Group) off() {
	Log.Infof("Group %s set to OFF", g.ID())

	g.Lock()
	old := g.state
	g.state = OFF
	g.Unlock()

	g.notifyListeners(old, OFF)
}

func (g *Group) refresh() {
	g.RLock()
	off := g.state == OFF
	g.RUnlock()

	for _, object := range g.objects {
		if object.State() != OFF {
			if off {
				g.on()
			}
			return
		}
	}

	if !off {
		g.off()
	}
}

func (g *Group) OnStateChange(object Object, old string, new string) {
	g.refresh()
}

// Add adds an Object to the group.
func (g *Group) Add(o Object) {
	g.objects = append(g.objects, o)

	o.AddObjectListener(g)

	g.refresh()
}

func (g *Group) Items() []Item {
	items := []Item{
		g.Item(),
	}
	for _, object := range g.objects {
		for _, item := range object.Items() {
			items = append(items, item)
		}
	}
	return items
}

func (gi *GroupItem) HTML() template.HTML {
	data := struct {
		ID    string
		Label string
		Img   string
		Items []Item
	}{
		ID:    gi.object.ID() + "_" + gi.ID(),
		Label: gi.object.Label(),
		Img:   fmt.Sprintf("statics/img/%s.png", gi.img),
	}

	for _, object := range gi.object.(*Group).objects {
		for _, item := range object.Items() {
			data.Items = append(data.Items, item)
		}
	}

	return itemTemplate("statics/items/group.html", data)
}

func (gi *GroupItem) MarshalJSON() ([]byte, error) {
	var items []Item

	for _, object := range gi.object.(*Group).objects {
		for _, item := range object.Items() {
			items = append(items, item)
		}
	}

	return json.Marshal(&struct {
		ID       string `json:"id"`
		ObjectID string `json:"oid"`
		Label    string `json:"label"`
		Value    string `json:"value"`
		Img      string `json:"img"`
		Items    []Item `json:"items"`
	}{
		ID:       gi.ID(),
		ObjectID: gi.Object().ID(),
		Label:    gi.Object().Label(),
		Value:    gi.Value(),
		Img:      gi.Img(),
		Items:    items,
	})
}

func newGroup(id string, label string) *Group {
	g := &Group{
		AnObject: AnObject{
			id:    id,
			label: label,
			items: make(map[string]Item),
		},
	}

	g.items[ItemID] = &GroupItem{
		AnItem: AnItem{
			object: g,
			img:    "group",
		},
	}

	return g
}

// RegisterGroup registers a new group with the given ID and label. Group
// is a collection of Objects. The state of the Group reflects the state of
// the hosted objects using a OR operator.
func RegisterGroup(id string, label string) *Group {
	g := newGroup(id, label)
	RegisterObject(g)
	return g
}
