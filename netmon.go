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
	"net"
	"sync/atomic"
	"time"

	fastping "github.com/tatsushid/go-fastping"
)

type NetMon struct {
	AnObject
	pinger      *fastping.Pinger
	pinging     atomic.Value
	lastSuccess time.Time
	retry       int
	fail        int
}

type NetMonItem struct {
	AnItem
}

const (
	pingTimeout = 5 * time.Second
)

func (n *NetMon) SetState(new string) string {
	switch new {
	case "on", "ON", "1":
		new = ON
	default:
		new = OFF
	}
	old := n.AnObject.SetState(new)

	if new != old {
		Log.Infof("NetMon set %s to %s", n.id, new)
	}

	if new != old {
		n.notifyListeners(old, new)
	}

	return old
}

func (n *NetMon) onRecv(addr *net.IPAddr, rtt time.Duration) {
	n.pinging.Store(false)

	n.lastSuccess = time.Now()
	n.fail = 0

	n.SetState(ON)
}

func (n *NetMon) onIdle() {
	n.pinging.Store(false)

	if n.lastSuccess.Add(pingTimeout).After(time.Now()) {
		return
	}

	n.fail++
	if n.fail > n.retry {
		n.SetState(OFF)
	} else {
		Log.Infof("NetMon error %s, fails: %d", n.id, n.fail)
	}
}

func (n *NetMon) refreshFnc() {
	if n.pinging.Load() == true {
		time.Sleep(time.Second)
		return
	}

	n.pinging.Store(true)
	if err := n.pinger.Run(); err != nil {
		Log.Fatalf("NetMon check error: %s", err)
		n.pinging.Store(false)
	}
}

func (n *NetMon) refresh(refresh time.Duration) {
	n.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		n.refreshFnc()
	}
}

func (ni *NetMonItem) HTML() template.HTML {
	data := struct {
		ID    string
		Label string
		Img   string
	}{
		ID:    ni.object.ID() + "_" + ni.ID(),
		Label: ni.object.Label(),
		Img:   fmt.Sprintf("statics/img/%s.png", ni.img),
	}

	return itemTemplate("statics/items/netmon.html", data)
}

func (ni *NetMonItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(ni)
}

func newNetMon(id string, label string, address string, refresh time.Duration, retry int) *NetMon {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", address)
	if err != nil {
		Log.Fatalf("NetMon address resolution error: %s", err)
	}
	p.AddIPAddr(ra)

	n := &NetMon{
		AnObject: AnObject{
			id:    id,
			label: label,
			items: make(map[string]Item),
		},
		pinger: p,
		retry:  retry,
	}

	n.items[ItemID] = &NetMonItem{
		AnItem: AnItem{
			object: n,
			img:    "netmon",
		},
	}

	p.OnRecv = n.onRecv
	p.OnIdle = n.onIdle

	go n.refresh(refresh)

	return n
}

// RegisterNetMon monitors pinging (ICMP Echo/Reply) the given address. It sends a ping every
// refresh delay and set the device as OFF after retry.
func RegisterNetMon(id string, label string, address string, refresh time.Duration, retry int) *NetMon {
	s := newNetMon(id, label, address, refresh, retry)
	RegisterObject(s)
	return s
}
