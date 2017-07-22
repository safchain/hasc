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
	"html/template"
	"sort"
	"sync"
)

const (
	ON  = "ON"
	OFF = "OFF"
)

type ObjectListener interface {
	OnStateChange(object Object, old string, new string)
}

type Item interface {
	ID() string
	Object() Object
	Value() string
	HTML() template.HTML
	SetImg(img string)
	Img() string
	Label() string
	MarshalJSON() ([]byte, error)
	Index() int
}

type AnItem struct {
	object Object
	img    string
	index  int
}

type Object interface {
	ID() string
	Label() string
	Address1() int
	Address2() int
	Receiver() int
	SetState(new string)
	State() string
	AddObjectListener(l ObjectListener)
	Items() []Item
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type AnObject struct {
	lock           sync.RWMutex
	id             string
	label          string
	address1       int
	address2       int
	receiver       int
	state          string
	eventListeners []ObjectListener
	items          map[string]Item
}

const (
	ItemID = "ITEM"
)

func (a *AnObject) Notify(l ObjectListener) {
	a.AddObjectListener(l)
}

func (a *AnObject) AddObjectListener(l ObjectListener) {
	for _, el := range a.eventListeners {
		if el == l {
			return
		}
	}
	a.eventListeners = append(a.eventListeners, l)
}

func (a *AnObject) notifyListeners(old string, new string) {
	for _, l := range a.eventListeners {
		l.OnStateChange(a, old, new)
	}
}

func (a *AnObject) ID() string {
	return a.id
}

func (a *AnObject) Label() string {
	return a.label
}

func (a *AnObject) Address1() int {
	return a.address1
}

func (a *AnObject) Address2() int {
	return a.address2
}

func (a *AnObject) Receiver() int {
	return a.receiver
}

func (a *AnObject) SetState(state string) {
	a.lock.Lock()
	a.state = state
	a.lock.Unlock()
}

func (a *AnObject) State() string {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.state
}

func (a *AnObject) HTML() template.HTML {
	return ""
}

type byIndex []Item

func (s byIndex) Len() int {
	return len(s)
}
func (s byIndex) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byIndex) Less(i, j int) bool {
	return s[i].Index() < s[j].Index()
}

func (a *AnObject) Items() []Item {
	var items []Item
	if a.items == nil {
		return items
	}

	for _, item := range a.items {
		items = append(items, item)
	}

	sort.Sort(byIndex(items))

	return items
}

func (a *AnObject) Item() Item {
	return a.items[ItemID]
}

func (a *AnObject) Lock() {
	a.lock.Lock()
}

func (a *AnObject) Unlock() {
	a.lock.Unlock()
}

func (a *AnObject) RLock() {
	a.lock.RLock()
}

func (a *AnObject) RUnlock() {
	a.lock.RUnlock()
}

func (ai *AnItem) SetImg(img string) {
	ai.img = img
}

func (ai *AnItem) Img() string {
	return ai.img
}

func (ai *AnItem) Index() int {
	return ai.index
}

func (ai *AnItem) Label() string {
	return ai.Object().Label()
}

func (ai *AnItem) Object() Object {
	return ai.object
}

func (ai *AnItem) Value() string {
	return ai.object.State()
}

func (ai *AnItem) ID() string {
	return ItemID
}

func marshalJSON(item Item) ([]byte, error) {
	return json.Marshal(&struct {
		ID       string `json:"id"`
		ObjectID string `json:"oid"`
		Label    string `json:"label"`
		Value    string `json:"value"`
		Img      string `json:"img"`
	}{
		ID:       item.ID(),
		ObjectID: item.Object().ID(),
		Label:    item.Label(),
		Value:    item.Value(),
		Img:      item.Img(),
	})
}
