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

package item

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

type CallbackListener struct {
	CbFnc func(item Item, old string, new string)
}

func (c *CallbackListener) OnValueChange(item Item, old string, new string) {
	c.CbFnc(item, old, new)
}

type Item interface {
	GetID() string
	GetType() string
	GetValue() string
	SetValue(string) (string, bool)
	GetImg() string
	SetImg(img string)
	GetLabel() string
	SetLabel(label string)
	GetUnit() string
	SetUnit(unit string)
	GetLastValueUpdate() time.Time

	AddListener(l ItemListener)
	MarshalJSON() ([]byte, error)
}

type AnItem struct {
	barrier int64
	lock    sync.RWMutex

	ID    string
	Label string
	Type  string
	Img   string
	Unit  string

	value           string
	lastValueUpdate time.Time
	lastValueChange time.Time

	listeners []ItemListener
	_         uint8
}

func (a *AnItem) AddListener(l ItemListener) {
	for _, el := range a.listeners {
		if el == l {
			return
		}
	}
	a.listeners = append(a.listeners, l)
}

func (a *AnItem) notifyListeners(old string, new string) {
	if atomic.CompareAndSwapInt64(&a.barrier, 0, 1) {
		for _, l := range a.listeners {
			l.OnValueChange(a, old, new)
		}
		atomic.StoreInt64(&a.barrier, 0)
	}
}

func (a *AnItem) GetID() string {
	return a.ID
}

func (a *AnItem) SetLabel(label string) {
	a.Label = label
}

func (a *AnItem) GetLabel() string {
	return a.Label
}

func (a *AnItem) SetValue(value string) (string, bool) {
	var (
		updated  bool
		oldValue string
	)

	a.lock.Lock()
	oldValue, a.value = a.value, value
	a.lastValueUpdate = time.Now()

	if oldValue != a.value {
		a.lastValueChange = a.lastValueUpdate
	}
	a.lock.Unlock()

	a.notifyListeners(oldValue, a.value)

	return oldValue, updated
}

func (a *AnItem) GetValue() string {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.value
}

func (a *AnItem) GetLastValueUpdate() time.Time {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.lastValueUpdate
}

func (a *AnItem) GetLastValueChange() time.Time {
	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.lastValueChange
}

func (a *AnItem) SetImg(img string) {
	a.Img = img
}

func (a *AnItem) GetImg() string {
	return a.Img
}

func (a *AnItem) GetType() string {
	return a.Type
}

func (a *AnItem) SetUnit(unit string) {
	a.Unit = unit
}

func (a *AnItem) GetUnit() string {
	return a.Unit
}

func (a *AnItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(a)
}

func marshalJSON(item Item) ([]byte, error) {
	const layout = "15:04:05"
	var lastUpdate string

	if !item.GetLastValueUpdate().IsZero() {
		lastUpdate = item.GetLastValueUpdate().Format(layout)
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
		ID:         item.GetID(),
		Type:       item.GetType(),
		Label:      item.GetLabel(),
		Value:      item.GetValue(),
		Img:        item.GetImg(),
		Unit:       item.GetUnit(),
		LastUpdate: lastUpdate,
	})
}
