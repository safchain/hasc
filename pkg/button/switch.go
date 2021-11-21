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

package button

import (
	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type SwitchItem struct {
	item.AnItem
}

func (s *SwitchItem) SetValue(value string) (string, bool) {
	switch value {
	case "on", "ON", "1":
		return s.AnItem.SetValue(item.ON)
	}

	return s.AnItem.SetValue(item.OFF)
}

func NewSwitchItem(id string, label string, disabled bool) *SwitchItem {
	kind := "switch"
	if disabled {
		kind = "state"
	}

	s := &SwitchItem{
		AnItem: item.AnItem{
			ID:    id,
			Label: label,
			Type:  kind,
			Img:   "switch",
		},
	}
	s.AnItem.SetValue(item.OFF)

	server.Registry.Add(s, "")

	return s
}
