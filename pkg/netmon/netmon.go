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

package netmon

import (
	"net"
	"sync/atomic"
	"time"

	fastping "github.com/tatsushid/go-fastping"

	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type NetMonItem struct {
	item.AnItem

	pinger      *fastping.Pinger
	pinging     atomic.Value
	lastSuccess time.Time
	retry       int
	fail        int
}

const (
	pingTimeout = 5 * time.Second
)

func (n *NetMonItem) SetValue(new string) (string, bool) {
	switch new {
	case "on", "ON", "1":
		new = item.ON
	default:
		new = item.OFF
	}

	return n.AnItem.SetValue(new)
}

func (n *NetMonItem) onRecv(addr *net.IPAddr, rtt time.Duration) {
	n.pinging.Store(false)

	n.lastSuccess = time.Now()
	n.fail = 0

	n.SetValue(item.ON)
}

func (n *NetMonItem) onIdle() {
	n.pinging.Store(false)

	if n.lastSuccess.Add(pingTimeout).After(time.Now()) {
		return
	}

	n.fail++
	if n.fail > n.retry {
		n.SetValue(item.OFF)
	} else {
		server.Log.Infof("NetMon error %s, fails: %d", n.GetID(), n.fail)
	}
}

func (n *NetMonItem) refreshFnc() {
	if n.pinging.Load() == true {
		time.Sleep(time.Second)
		return
	}

	n.pinging.Store(true)
	if err := n.pinger.Run(); err != nil {
		server.Log.Fatalf("NetMon check error: %s", err)
		n.pinging.Store(false)
	}
}

func (n *NetMonItem) refresh(refresh time.Duration) {
	n.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		n.refreshFnc()
	}
}

func NewNetMonItem(id string, label string, address string, refresh time.Duration, retry int) *NetMonItem {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", address)
	if err != nil {
		server.Log.Fatalf("NetMon address resolution error: %s", err)
	}
	p.AddIPAddr(ra)

	n := &NetMonItem{
		AnItem: item.AnItem{
			ID:    id,
			Label: label,
			Type:  "state",
			Img:   "netmon",
		},
		pinger: p,
		retry:  retry,
	}

	p.OnRecv = n.onRecv
	p.OnIdle = n.onIdle

	go n.refresh(refresh)

	server.Registry.Add(n, "")

	return n
}
