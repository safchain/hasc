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

package timer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type TimerItem struct {
	sync.RWMutex

	item.AnItem

	Item item.Item

	opts       TimerOpts
	lastUpdate time.Time
	active     atomic.Value
}

type TimerOpts struct {
	OnAfter  time.Duration
	OffAfter time.Duration
	Timeout  time.Duration
	OnState  string
	OffState string
}

func (r *TimerItem) on() (string, bool) {
	server.Log.Infof("Timer %s set to ON", r.GetID())

	old, _ := r.AnItem.SetValue(item.ON)

	if r.Item.GetValue() == item.ON {
		r.Item.SetValue(item.ON)
	}

	r.Lock()
	r.lastUpdate = time.Now()
	r.Unlock()

	if old != item.ON {
		if r.active.Load() == true {
			return old, true
		}

		go func() {
			defer func() {
				r.off()
				r.active.Store(false)
			}()

			r.active.Store(true)

			onAfter := time.After(r.opts.OnAfter)
			tick := time.NewTicker(time.Second)
			defer tick.Stop()

			var on bool
			for r.active.Load() == true {
				select {
				case <-onAfter:
					if r.Item != nil {
						r.Item.SetValue(r.opts.OnState)
					}
					on = true
				case now := <-tick.C:
					r.RLock()
					tm := r.lastUpdate.Add(r.opts.Timeout).Before(now)
					r.RUnlock()
					if tm {
						if on {
							// reset it as on as another timer could act on the same switch
							if r.Item.GetValue() != r.opts.OnState {
								r.Item.SetValue(r.opts.OnState)
							}

							r.RLock()
							diff := r.lastUpdate.Add(r.opts.OffAfter).Sub(now)
							r.RUnlock()
							if diff < 0 {
								return
							}

							remain := fmt.Sprintf("%d", int(diff.Seconds()))

							r.AnItem.SetValue(remain)
						} else {
							return
						}
					}
				}
			}
		}()
	}

	return old, false
}

func (r *TimerItem) off() (string, bool) {
	server.Log.Infof("Timer %s set to OFF", r.GetID())

	old, updated := r.AnItem.SetValue(item.OFF)

	if r.Item != nil {
		r.Item.SetValue(r.opts.OffState)
	}

	return old, updated
}

func (r *TimerItem) SetValue(new string) (string, bool) {
	switch new {
	case "on", "ON", "1":
		return r.on()
	}

	return r.off()
}

func NewTimerItem(id string, label string, it item.Item, opts ...TimerOpts) *TimerItem {
	r := &TimerItem{
		AnItem: item.AnItem{
			ID:    id,
			Label: label,
			Type:  "timer",
			Img:   "timer",
		},
		Item: it,
	}
	r.AnItem.SetValue(item.OFF)

	if len(opts) > 0 {
		r.opts = opts[0]
	}

	if r.opts.OnState == "" {
		r.opts.OnState = item.ON
	}
	if r.opts.OffState == "" {
		r.opts.OffState = item.OFF
	}
	if r.opts.Timeout == 0 {
		r.opts.Timeout = time.Second
	}

	server.Registry.Add(r, "")

	return r
}
