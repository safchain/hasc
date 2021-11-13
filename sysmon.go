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
	"math"
	"time"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type SysMon struct {
	memPercent AnItem
	cpuAvg1    AnItem
	cpuAvg5    AnItem
	cpuAvg15   AnItem
	uptime     AnItem
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
	Log.Infof("SysMon refresh")

	v, err := mem.VirtualMemory()
	if err != nil {
		Log.Errorf("SysMon memory refresh %s", err)
		return
	}

	if old, updated := s.memPercent.SetValue(fmt.Sprintf("%.2f", v.UsedPercent)); updated {
		s.memPercent.notifyListeners(old, s.memPercent.value)
	}

	l, err := load.Avg()
	if err != nil {
		Log.Errorf("SysMon cpu avg refresh %s", err)
		return
	}
	if old, updated := s.cpuAvg1.SetValue(fmt.Sprintf("%.2f", l.Load1)); updated {
		s.cpuAvg1.notifyListeners(old, s.cpuAvg1.value)
	}
	if old, updated := s.cpuAvg5.SetValue(fmt.Sprintf("%.2f", l.Load5)); updated {
		s.cpuAvg5.notifyListeners(old, s.cpuAvg5.value)
	}
	if old, updated := s.cpuAvg15.SetValue(fmt.Sprintf("%.2f", l.Load15)); updated {
		s.cpuAvg15.notifyListeners(old, s.cpuAvg15.value)
	}

	u, err := host.Uptime()
	if err != nil {
		Log.Errorf("SysMon uptime refresh %s", err)
		return
	}
	if old, updated := s.uptime.SetValue(secondsToHuman(u)); updated {
		s.cpuAvg15.notifyListeners(old, s.uptime.value)
	}
}

func (s *SysMon) refresh(refresh time.Duration) {
	s.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		s.refreshFnc()
	}
}

func (s *SysMon) MemUsedPercentItem() Item {
	return &s.memPercent
}

func (s *SysMon) CPUAvg1Item() Item {
	return &s.cpuAvg1
}

func (s *SysMon) CPUAvg5Item() Item {
	return &s.cpuAvg5
}

func (s *SysMon) CPUAvg15Item() Item {
	return &s.cpuAvg15
}

func (s *SysMon) UptimeItem() Item {
	return &s.uptime
}

func NewSysMon(id string, label string, refresh time.Duration) *SysMon {
	s := &SysMon{
		memPercent: AnItem{
			id:    id + "_MEM",
			label: "Mem. used",
			kind:  "value",
			img:   "mem",
			unit:  "%",
		},
		cpuAvg1: AnItem{
			id:    id + "_AVG1",
			label: "CPU Avg1",
			kind:  "value",
			img:   "cpu",
			unit:  "%",
		},
		cpuAvg5: AnItem{
			id:    id + "_AVG5",
			label: "CPU Avg5",
			kind:  "value",
			img:   "cpu",
			unit:  "%",
		},
		cpuAvg15: AnItem{
			id:    id + "_AVG15",
			label: "CPU Avg15",
			kind:  "value",
			img:   "cpu",
			unit:  "%",
		},
		uptime: AnItem{
			id:    id + "_UPTIME",
			label: "Uptime",
			kind:  "value",
			img:   "clock",
		},
	}

	// first init to retrieve all the items
	s.refreshFnc()

	go s.refresh(refresh)

	registry.Add(&s.memPercent)
	registry.Add(&s.cpuAvg1)
	registry.Add(&s.cpuAvg5)
	registry.Add(&s.cpuAvg15)
	registry.Add(&s.uptime)

	return s
}
