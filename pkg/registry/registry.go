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

package registry

import (
	"fmt"
	"sync"

	"github.com/safchain/hasc/pkg/item"
)

type Registry struct {
	sync.RWMutex

	items     map[string]item.Item
	listeners []item.ItemListener
}

func (r *Registry) AddListener(l item.ItemListener) {
	for _, el := range r.listeners {
		if el == l {
			return
		}
	}
	r.listeners = append(r.listeners, l)
}

func (r *Registry) Add(it item.Item, prefix string) {
	r.Lock()
	defer r.Unlock()

	key := it.GetID()
	if prefix != "" {
		key = fmt.Sprintf("%s/%s", prefix, key)
	}
	r.items[key] = it

	for _, l := range r.listeners {
		it.AddListener(l)
	}
}

func (r *Registry) Get(id string) item.Item {
	r.RLock()
	defer r.RUnlock()

	return r.items[id]
}

func (r *Registry) Items() []item.Item {
	var items []item.Item

	for _, it := range r.items {
		items = append(items, it)
	}

	return items
}

func NewRegistry(listeners ...item.ItemListener) *Registry {
	return &Registry{
		items:     make(map[string]item.Item),
		listeners: listeners,
	}
}
