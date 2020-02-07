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
	"net"
	"sync/atomic"
	"time"

	fastping "github.com/tatsushid/go-fastping"
)

type NetMon struct {
	AnItem
	pinger      *fastping.Pinger
	pinging     atomic.Value
	lastSuccess time.Time
	retry       int
	fail        int
}

const (
	pingTimeout = 5 * time.Second
)

func (n *NetMon) SetValue(new string) (string, bool) {
	switch new {
	case "on", "ON", "1":
		new = ON
	default:
		new = OFF
	}

	old, updated := n.AnItem.SetValue(new)
	if updated {
		n.notifyListeners(old, new)
	}

	return old, updated
}

func (n *NetMon) onRecv(addr *net.IPAddr, rtt time.Duration) {
	n.pinging.Store(false)

	n.lastSuccess = time.Now()
	n.fail = 0

	n.SetValue(ON)
}

func (n *NetMon) onIdle() {
	n.pinging.Store(false)

	if n.lastSuccess.Add(pingTimeout).After(time.Now()) {
		return
	}

	n.fail++
	if n.fail > n.retry {
		n.SetValue(OFF)
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

func NewNetMon(id string, label string, address string, refresh time.Duration, retry int) *NetMon {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", address)
	if err != nil {
		Log.Fatalf("NetMon address resolution error: %s", err)
	}
	p.AddIPAddr(ra)

	n := &NetMon{
		AnItem: AnItem{
			id:    id,
			label: label,
			kind:  "state",
			img:   "netmon",
		},
		pinger: p,
		retry:  retry,
	}

	p.OnRecv = n.onRecv
	p.OnIdle = n.onIdle

	go n.refresh(refresh)

	registry.Add(n)

	return n
}
