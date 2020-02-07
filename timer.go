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
	"fmt"
	"sync/atomic"
	"time"
)

type Timer struct {
	AnItem
	item       Item
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

type TimerItem struct {
	AnItem
}

func (r *Timer) on() (string, bool) {
	Log.Infof("Timer %s set to ON", r.ID())

	old, _ := r.AnItem.SetValue(ON)
	r.notifyListeners(old, ON)

	if r.item.Value() == ON {
		r.item.SetValue(ON)
	}

	r.Lock()
	r.lastUpdate = time.Now()
	r.Unlock()

	if old != ON {
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
					if r.item != nil {
						r.item.SetValue(r.opts.OnState)
					}
					on = true
				case now := <-tick.C:
					r.RLock()
					tm := r.lastUpdate.Add(r.opts.Timeout).Before(now)
					r.RUnlock()
					if tm {
						if on {
							// reset it as on as another timer could act on the same switch
							if r.item.Value() != r.opts.OnState {
								r.item.SetValue(r.opts.OnState)
							}

							r.RLock()
							diff := r.lastUpdate.Add(r.opts.OffAfter).Sub(now)
							r.RUnlock()
							if diff < 0 {
								return
							}

							remain := fmt.Sprintf("%d", int(diff.Seconds()))

							old, _ := r.AnItem.SetValue(remain)

							r.notifyListeners(old, remain)
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

func (r *Timer) off() (string, bool) {
	Log.Infof("Timer %s set to OFF", r.ID())

	old, updated := r.AnItem.SetValue(OFF)
	r.notifyListeners(old, OFF)

	if r.item != nil {
		r.item.SetValue(r.opts.OffState)
	}

	return old, updated
}

func (r *Timer) SetValue(new string) (string, bool) {
	var old string
	var updated bool

	switch new {
	case "on", "ON", "1":
		old, updated = r.on()
	default:
		old, updated = r.off()
	}

	return old, updated
}

func NewTimer(id string, label string, item Item, opts ...TimerOpts) *Timer {
	r := &Timer{
		AnItem: AnItem{
			id:    id,
			label: label,
			value: OFF,
			kind:  "timer",
			img:   "timer",
		},
		item: item,
	}
	if len(opts) > 0 {
		r.opts = opts[0]
	}

	if r.opts.OnState == "" {
		r.opts.OnState = ON
	}
	if r.opts.OffState == "" {
		r.opts.OffState = OFF
	}
	if r.opts.Timeout == 0 {
		r.opts.Timeout = time.Second
	}

	registry.Add(r)

	return r
}
