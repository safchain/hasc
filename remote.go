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
	"sync"
	"sync/atomic"
	"time"
)

type Remote struct {
	AnObject
	obj        Object
	opts       RemoteOpts
	lastUpdate time.Time
	active     atomic.Value
	wg         sync.WaitGroup
}

type RemoteOpts struct {
	OnAfter  time.Duration
	OffAfter time.Duration
	Timeout  time.Duration
	OnState  string
	OffState string
}

type RemoteItem struct {
	AnItem
}

func (r *Remote) on() {
	Log.Infof("Remote %s set to ON", r.ID())

	r.Lock()
	old := r.state
	r.lastUpdate = time.Now()
	r.Unlock()

	if old != ON {
		if r.active.Load() == true {
			r.active.Store(false)
			r.wg.Wait()
		}

		r.Lock()
		r.state = ON
		r.Unlock()

		r.notifyListeners(old, ON)

		r.wg.Add(1)
		go func() {
			defer func() {
				r.off()

				r.active.Store(false)
				r.wg.Done()
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
}

func (r *Remote) off() {
	Log.Infof("Remote %s set to OFF", r.ID())

	r.Lock()
	old := r.state
	r.state = OFF
	r.Unlock()

	if old != OFF {
		r.notifyListeners(old, OFF)
	}

	r.obj.SetState(r.opts.OffState)
}

func (r *Remote) SetState(new string) {
	switch new {
	case "on", "ON", "1":
		r.on()
	default:
		r.off()
	}
}

func (ri *RemoteItem) HTML() template.HTML {
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

	return itemTemplate("statics/items/remote.html", data)
}

func (ri *RemoteItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(ri)
}

func newRemote(id string, label string, address1 int, address2 int, receiver int, obj Object, opts ...RemoteOpts) *Remote {
	r := &Remote{
		AnObject: AnObject{
			id:       id,
			label:    label,
			address1: address1,
			address2: address2,
			receiver: receiver,
			items:    make(map[string]Item),
			state:    OFF,
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

	r.items[ItemID] = &RemoteItem{
		AnItem: AnItem{
			object: r,
			img:    "switch",
		},
	}

	return r
}

// RegisterRemote triggers state change on the given object. It can act as a timer meaning
// it can set the object state to ON after a certain among of time and can delay the OFF
// state.
func RegisterRemote(id string, label string, address1 int, address2 int, receiver int, obj Object, opts ...RemoteOpts) *Remote {
	r := newRemote(id, label, address1, address2, receiver, obj, opts...)
	RegisterObject(r)
	return r
}
