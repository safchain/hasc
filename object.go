/*
 * Copyright (C) 2020 Sylvain Afchain
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
	"sync"
	"sync/atomic"
	"time"
)

const (
	ON  = "ON"
	OFF = "OFF"
)

type ItemListener interface {
	OnValueChange(item Item, old string, new string)
}

type Item interface {
	ID() string
	Type() string
	SetValue(string) (string, bool)
	Value() string
	SetImg(img string)
	Img() string
	SetLabel(label string)
	Label() string
	SetUnit(unit string)
	Unit() string
	LastValueUpdate() time.Time
	AddListener(l ItemListener)
	MarshalJSON() ([]byte, error)
}

type AnItem struct {
	barrier         int64
	lock            sync.RWMutex
	id              string
	label           string
	eventListeners  []ItemListener
	kind            string
	img             string
	unit            string
	value           string
	lastValueUpdate time.Time
}

func (a *AnItem) AddListener(l ItemListener) {
	for _, el := range a.eventListeners {
		if el == l {
			return
		}
	}
	a.eventListeners = append(a.eventListeners, l)
}

func (a *AnItem) notifyListeners(old string, new string) {
	if atomic.CompareAndSwapInt64(&a.barrier, 0, 1) {
		for _, l := range a.eventListeners {
			l.OnValueChange(a, old, new)
		}
		atomic.StoreInt64(&a.barrier, 0)
	}
}

func (a *AnItem) ID() string {
	return a.id
}

func (a *AnItem) SetLabel(label string) {
	a.label = label
}

func (a *AnItem) Label() string {
	return a.label
}

func (a *AnItem) SetValue(value string) (string, bool) {
	updated := false
	var old string

	a.lock.Lock()
	if value != a.value {
		updated = true
	}
	old = a.value
	a.value = value
	a.lastValueUpdate = time.Now()
	a.lock.Unlock()

	if updated {
		Log.Infof("%s set to %s", a.ID(), value)
	}

	return old, updated
}

func (a *AnItem) Value() string {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.value
}

func (a *AnItem) LastValueUpdate() time.Time {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.lastValueUpdate
}

func (a *AnItem) Lock() {
	a.lock.Lock()
}

func (a *AnItem) Unlock() {
	a.lock.Unlock()
}

func (a *AnItem) RLock() {
	a.lock.RLock()
}

func (a *AnItem) RUnlock() {
	a.lock.RUnlock()
}

func (a *AnItem) SetImg(img string) {
	a.img = img
}

func (a *AnItem) Img() string {
	return a.img
}

func (a *AnItem) Type() string {
	return a.kind
}

func (a *AnItem) SetUnit(unit string) {
	a.unit = unit
}

func (a *AnItem) Unit() string {
	return a.unit
}

func (a *AnItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(a)
}

func marshalJSON(item Item) ([]byte, error) {
	const layout = "15:04:05"
	var lastUpdate string

	if !item.LastValueUpdate().IsZero() {
		lastUpdate = item.LastValueUpdate().Format(layout)
	}

	return json.Marshal(&struct {
		ID         string
		Type       string
		Label      string
		Value      string
		Img        string
		Unit       string
		LastUpdate string
	}{
		ID:         item.ID(),
		Type:       item.Type(),
		Label:      item.Label(),
		Value:      item.Value(),
		Img:        item.Img(),
		Unit:       item.Unit(),
		LastUpdate: lastUpdate,
	})
}
