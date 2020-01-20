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
	"sort"
	"sync"
	"sync/atomic"
	"time"
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
	SetImg(img string)
	Img() string
	Label() string
	MarshalJSON() ([]byte, error)
	Index() int
	Type() string
	Unit() string
}

type AnItem struct {
	object Object
	kind   string
	img    string
	index  int
	unit   string
}

type Object interface {
	ID() string
	Label() string
	Device() interface{}
	SetState(new string) string
	State() string
	AddObjectListener(l ObjectListener)
	Items() []Item
	Lock()
	Unlock()
	RLock()
	RUnlock()
	LastStateUpdate() time.Time
}

type AnObject struct {
	lock            sync.RWMutex
	id              string
	label           string
	device          interface{}
	state           string
	eventListeners  []ObjectListener
	items           map[string]Item
	barrier         int64
	lastStateUpdate time.Time
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
	if atomic.CompareAndSwapInt64(&a.barrier, 0, 1) {
		for _, l := range a.eventListeners {
			l.OnStateChange(a, old, new)
		}
		atomic.StoreInt64(&a.barrier, 0)
	}
}

func (a *AnObject) ID() string {
	return a.id
}

func (a *AnObject) Label() string {
	return a.label
}

func (a *AnObject) Device() interface{} {
	return a.device
}

func (a *AnObject) SetState(state string) string {
	a.lock.Lock()
	old := a.state
	a.state = state
	a.lastStateUpdate = time.Now()
	a.lock.Unlock()

	return old
}

func (a *AnObject) LastStateUpdate() time.Time {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.lastStateUpdate
}

func (a *AnObject) State() string {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.state
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

func (ai *AnItem) Type() string {
	return ai.kind
}

func (ai *AnItem) Unit() string {
	return ai.unit
}

func marshalJSON(item Item) ([]byte, error) {
	const layout = "15:04:05"
	var lastUpdate string

	if !item.Object().LastStateUpdate().IsZero() {
		lastUpdate = item.Object().LastStateUpdate().Format(layout)
	}

	return json.Marshal(&struct {
		ID         string `json:"id"`
		ObjectID   string `json:"oid"`
		Type       string `json:"type"`
		Label      string `json:"label"`
		Value      string `json:"value"`
		Img        string `json:"img"`
		Unit       string `json:"unit"`
		LastUpdate string `json:"lastupdate"`
	}{
		ID:         item.ID(),
		ObjectID:   item.Object().ID(),
		Type:       item.Type(),
		Label:      item.Label(),
		Value:      item.Value(),
		Img:        item.Img(),
		Unit:       item.Unit(),
		LastUpdate: lastUpdate,
	})
}
