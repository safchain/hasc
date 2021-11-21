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

package sysmon

import (
	"fmt"
	"math"
	"time"

	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type SysMon struct {
	MemPercentItem *item.AnItem
	CPUAvg1Item    *item.AnItem
	CPUAvg5Item    *item.AnItem
	CPUAvg15Item   *item.AnItem
	UptimeItem     *item.AnItem
}

func secondsToHuman(input uint64) (result string) {
	years := int(math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12))
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := int(math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30))
	seconds = input % (60 * 60 * 24 * 7)
	days := int(math.Floor(float64(seconds) / 60 / 60 / 24))
	seconds = input % (60 * 60 * 24)
	hours := int(math.Floor(float64(seconds) / 60 / 60))
	seconds = input % (60 * 60)
	minutes := int(math.Floor(float64(seconds) / 60))
	seconds = input % 60

	if years > 0 {
		result = fmt.Sprintf("%dY %dM %dd %dh %dm %ds", years, months, days, hours, minutes, seconds)
	} else if months > 0 {
		result = fmt.Sprintf("%dM %dd %dh %dm %ds", months, days, hours, minutes, seconds)
	} else if days > 0 {
		result = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		result = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		result = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		result = fmt.Sprintf("%ds", seconds)
	}

	return
}

func (s *SysMon) refreshFnc() {
	server.Log.Infof("SysMon refresh")

	v, err := mem.VirtualMemory()
	if err != nil {
		server.Log.Errorf("SysMon memory refresh %s", err)
		return
	}

	s.MemPercentItem.SetValue(fmt.Sprintf("%.2f", v.UsedPercent))

	l, err := load.Avg()
	if err != nil {
		server.Log.Errorf("SysMon cpu avg refresh %s", err)
		return
	}
	s.CPUAvg1Item.SetValue(fmt.Sprintf("%.2f", l.Load1))
	s.CPUAvg5Item.SetValue(fmt.Sprintf("%.2f", l.Load5))
	s.CPUAvg15Item.SetValue(fmt.Sprintf("%.2f", l.Load15))

	u, err := host.Uptime()
	if err != nil {
		server.Log.Errorf("SysMon uptime refresh %s", err)
		return
	}
	s.UptimeItem.SetValue(secondsToHuman(u))
}

func (s *SysMon) refresh(refresh time.Duration) {
	s.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		s.refreshFnc()
	}
}

func NewSysMon(id string, label string, refresh time.Duration) *SysMon {
	s := &SysMon{
		MemPercentItem: &item.AnItem{
			ID:    id + "_MEM",
			Label: "Mem. used",
			Type:  "value",
			Img:   "mem",
			Unit:  "%",
		},
		CPUAvg1Item: &item.AnItem{
			ID:    id + "_AVG1",
			Label: "CPU Avg1",
			Type:  "value",
			Img:   "cpu",
			Unit:  "%",
		},
		CPUAvg5Item: &item.AnItem{
			ID:    id + "_AVG5",
			Label: "CPU Avg5",
			Type:  "value",
			Img:   "cpu",
			Unit:  "%",
		},
		CPUAvg15Item: &item.AnItem{
			ID:    id + "_AVG15",
			Label: "CPU Avg15",
			Type:  "value",
			Img:   "cpu",
			Unit:  "%",
		},
		UptimeItem: &item.AnItem{
			ID:    id + "_UPTIME",
			Label: "Uptime",
			Type:  "value",
			Img:   "clock",
		},
	}

	// first init to retrieve all the items
	s.refreshFnc()

	go s.refresh(refresh)

	server.Registry.Add(s.MemPercentItem, id)
	server.Registry.Add(s.CPUAvg1Item, id)
	server.Registry.Add(s.CPUAvg5Item, id)
	server.Registry.Add(s.CPUAvg15Item, id)
	server.Registry.Add(s.UptimeItem, id)

	return s
}
