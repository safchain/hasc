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
	"html/template"
	"sync/atomic"
	"time"
)

type Timer struct {
	AnObject
	obj        Object
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

func (r *Timer) on() string {
	Log.Infof("Timer %s set to ON", r.ID())

	old := r.AnObject.SetState(ON)

	r.Lock()
	r.lastUpdate = time.Now()
	r.Unlock()

	if old != ON {
		r.notifyListeners(old, ON)

		if r.active.Load() == true {
			if r.obj.State() != r.opts.OnState {
				r.obj.SetState(r.opts.OnState)
			}
			return old
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
					r.obj.SetState(r.opts.OnState)
					on = true
				case now := <-tick.C:
					r.RLock()
					tm := r.lastUpdate.Add(r.opts.Timeout).Before(now)
					r.RUnlock()
					if tm {
						if on {
							// reset it as on as another timer could act on the same switch
							if r.obj.State() != r.opts.OnState {
								r.obj.SetState(r.opts.OnState)
							}

							r.RLock()
							diff := r.lastUpdate.Add(r.opts.OffAfter).Sub(now)
							r.RUnlock()
							if diff < 0 {
								return
							}

							remain := fmt.Sprintf("%d", int(diff.Seconds()))
							r.Lock()
							old := r.state
							r.state = remain
							r.Unlock()

							r.notifyListeners(old, remain)
						} else {
							return
						}
					}
				}
			}
		}()
	}

	return old
}

func (r *Timer) off() string {
	Log.Infof("Timer %s set to OFF", r.ID())

	old := r.AnObject.SetState(OFF)

	if old != OFF {
		r.notifyListeners(old, OFF)
	}

	r.obj.SetState(r.opts.OffState)

	return old
}

func (r *Timer) SetState(new string) string {
	var old string
	switch new {
	case "on", "ON", "1":
		old = r.on()
	default:
		old = r.off()
	}

	return old
}

func (ri *TimerItem) HTML() template.HTML {
	data := struct {
		ID       string
		ObjectID string
		Label    string
		Img      string
	}{
		ID:       ri.object.ID() + "_" + ri.ID(),
		ObjectID: ri.object.ID(),
		Label:    ri.object.Label(),
		Img:      fmt.Sprintf("statics/img/%s.png", ri.img),
	}

	return itemTemplate("statics/items/timer.html", data)
}

func (ri *TimerItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(ri)
}

func newTimer(id string, label string, device interface{}, obj Object, opts ...TimerOpts) *Timer {
	r := &Timer{
		AnObject: AnObject{
			id:     id,
			label:  label,
			device: device,
			items:  make(map[string]Item),
			state:  OFF,
		},
		obj: obj,
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

	r.items[ItemID] = &TimerItem{
		AnItem: AnItem{
			object: r,
			img:    "switch",
		},
	}

	return r
}

// RegisterTimer triggers state change on the given object. It can act as a timer meaning
// it can set the object state to ON after a certain among of time and can delay the OFF
// state.
func RegisterTimer(id string, label string, device interface{}, obj Object, opts ...TimerOpts) *Timer {
	r := newTimer(id, label, device, obj, opts...)
	RegisterObject(r)
	return r
}
