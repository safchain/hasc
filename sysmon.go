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
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type diskValues struct {
	Name       string
	ReadBytes  uint64
	WriteBytes uint64
}

type sysMonValues struct {
	MemUsedPercent float64
	CPUAvg1        float64
	CPUAvg5        float64
	CPUAvg15       float64
	Uptime         uint64
	Disks          map[string]diskValues
}

type SysMon struct {
	AnObject
	values sysMonValues
}

type SysMonUint64Item struct {
	AnItem
	id    string
	label string
	value uint64
	unit  string
}

type SysMonFloat64Item struct {
	AnItem
	id    string
	label string
	value float64
	unit  string
}

const (
	MemUsedPercentID = "MEM_USED_PERCENT"
	CPUAvg1ID        = "CPU_AVG1"
	CPUAvg5ID        = "CPU_AVG5"
	CPUAvg15ID       = "CPU_AVG15"
)

func (s *SysMon) setState(new string) {
	Log.Infof("SysMon %s set to %s", s.ID(), new)

	s.Lock()
	old := s.state
	s.state = new
	s.Unlock()

	if new != old {
		s.notifyListeners(old, new)
	}
}

func (s *SysMon) refreshFnc() {
	Log.Infof("SysMon %s refresh", s.ID())
	v, err := mem.VirtualMemory()
	if err != nil {
		Log.Errorf("SysMon memory refresh %s", err)
		return
	}
	s.values.MemUsedPercent = v.UsedPercent
	s.items[MemUsedPercentID].(*SysMonFloat64Item).value = v.UsedPercent

	l, err := load.Avg()
	if err != nil {
		Log.Errorf("SysMon cpu avg refresh %s", err)
		return
	}
	s.values.CPUAvg1 = l.Load1
	s.items[CPUAvg1ID].(*SysMonFloat64Item).value = l.Load1
	s.values.CPUAvg5 = l.Load5
	s.items[CPUAvg5ID].(*SysMonFloat64Item).value = l.Load5
	s.values.CPUAvg15 = l.Load15
	s.items[CPUAvg15ID].(*SysMonFloat64Item).value = l.Load15

	u, err := host.Uptime()
	if err != nil {
		Log.Errorf("SysMon uptime refresh %s", err)
		return
	}
	s.values.Uptime = u

	d, err := disk.IOCounters()
	if err != nil {
		Log.Errorf("SysMon uptime refresh %s", err)
		return
	}

	index := 99
	for name, values := range d {
		s.values.Disks[name] = diskValues{
			Name:       name,
			ReadBytes:  values.ReadBytes,
			WriteBytes: values.WriteBytes,
		}

		id := fmt.Sprintf("DISK_%s_READ_BYTES", strings.ToUpper(name))
		if item, ok := s.items[id]; ok {
			item.(*SysMonUint64Item).value = values.ReadBytes
		} else {
			s.items[id] = &SysMonUint64Item{
				AnItem: AnItem{
					object: s,
					img:    "hdd",
					index:  index,
				},
				id:    id,
				label: fmt.Sprintf("%s read", name),
				value: values.ReadBytes,
			}
			index++
		}

		id = fmt.Sprintf("DISK_%s_WRITE_BYTES", strings.ToUpper(name))
		if item, ok := s.items[id]; ok {
			item.(*SysMonUint64Item).value = values.WriteBytes
		} else {
			s.items[id] = &SysMonUint64Item{
				AnItem: AnItem{
					object: s,
					img:    "hdd",
					index:  index,
				},
				id:    id,
				label: fmt.Sprintf("%s write", name),
				value: values.WriteBytes,
			}
			index++
		}
	}

	data, _ := json.Marshal(s.values)
	s.setState(string(data))
}

func (s *SysMon) refresh(refresh time.Duration) {
	s.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		s.refreshFnc()
	}
}

func (s *SysMon) MemUsedPercentItem() Item {
	return s.items[MemUsedPercentID]
}

func (s *SysMon) CPUAvg1Item() Item {
	return s.items[CPUAvg1ID]
}

func (s *SysMon) CPUAvg5Item() Item {
	return s.items[CPUAvg5ID]
}

func (s *SysMon) CPUAvg15Item() Item {
	return s.items[CPUAvg15ID]
}

func (si *SysMonFloat64Item) ID() string {
	return si.id
}

func (si *SysMonFloat64Item) Value() string {
	si.object.RLock()
	defer si.object.RUnlock()

	return fmt.Sprintf("%.2f", si.value)
}

func (si *SysMonFloat64Item) HTML() template.HTML {
	si.object.RLock()
	defer si.object.RUnlock()

	return valueTemplate(si, si.Label(), si.unit, si.img)
}

func (si *SysMonFloat64Item) Label() string {
	return si.label
}

func (si *SysMonFloat64Item) MarshalJSON() ([]byte, error) {
	return marshalJSON(si)
}

func (s *SysMon) DiskReadItem(name string) Item {
	id := fmt.Sprintf("DISK_%s_READ_BYTES", strings.ToUpper(name))
	if item, ok := s.items[id]; ok {
		return item
	}
	return &SysMonUint64Item{
		AnItem: AnItem{
			object: s,
			img:    "hdd",
			index:  200,
		},
		id:    id,
		label: fmt.Sprintf("%s not found", name),
	}
}

func (s *SysMon) DiskWriteItem(name string) Item {
	id := fmt.Sprintf("DISK_%s_WRITE_BYTES", strings.ToUpper(name))
	if item, ok := s.items[id]; ok {
		return item
	}
	return &SysMonUint64Item{
		AnItem: AnItem{
			object: s,
			img:    "hdd",
			index:  200,
		},
		id:    id,
		label: fmt.Sprintf("%s not found", name),
	}
}

func (si *SysMonUint64Item) ID() string {
	return si.id
}

func (si *SysMonUint64Item) Value() string {
	si.object.RLock()
	defer si.object.RUnlock()

	return fmt.Sprintf("%d", si.value)
}

func (si *SysMonUint64Item) HTML() template.HTML {
	si.object.RLock()
	defer si.object.RUnlock()

	return valueTemplate(si, si.Label(), si.unit, si.img)
}

func (si *SysMonUint64Item) Label() string {
	return si.label
}

func (si *SysMonUint64Item) MarshalJSON() ([]byte, error) {
	return marshalJSON(si)
}

func newSysMon(id string, label string, refresh time.Duration) *SysMon {
	s := &SysMon{
		AnObject: AnObject{
			id:    id,
			label: label,
			items: make(map[string]Item),
		},
	}
	s.values.Disks = make(map[string]diskValues)

	s.items[MemUsedPercentID] = &SysMonFloat64Item{
		AnItem: AnItem{
			object: s,
			img:    "mem",
			index:  0,
		},
		id:    MemUsedPercentID,
		label: "Mem. used",
		unit:  "%",
	}

	s.items[CPUAvg1ID] = &SysMonFloat64Item{
		AnItem: AnItem{
			object: s,
			img:    "cpu",
			index:  1,
		},
		id:    CPUAvg1ID,
		label: "CPU Avg1",
	}

	s.items[CPUAvg5ID] = &SysMonFloat64Item{
		AnItem: AnItem{
			object: s,
			img:    "cpu",
			index:  2,
		},
		id:    CPUAvg5ID,
		label: "CPU Avg5",
	}

	s.items[CPUAvg15ID] = &SysMonFloat64Item{
		AnItem: AnItem{
			object: s,
			img:    "cpu",
			index:  3,
		},
		id:    CPUAvg15ID,
		label: "CPU Avg15",
	}

	// first init to retrieve all the items
	s.refreshFnc()

	go s.refresh(refresh)

	return s
}

// RegisterSysMon registers a system monitor object. It monitor CPU, Memory and Disk usage.
func RegisterSysMon(id string, label string, refresh time.Duration) *SysMon {
	s := newSysMon(id, label, refresh)
	RegisterObject(s)
	return s
}
