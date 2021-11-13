/*
 * Copyright (C) 2020 Sylvain Afchain
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
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SmartBulb Object
type SmartBulb struct {
	id       string
	pubTopic string
	subTopic string
	conn     *MQTTConn
	tick     *AnItem
	color    Item
}

type color struct {
	AnItem
}

func (c *color) SetValue(new string) (string, bool) {
	old, updated := c.AnItem.SetValue(new)
	c.notifyListeners(old, new)
	return old, updated
}

func (s *SmartBulb) OnValueChange(item Item, old string, new string) {
	s.conn.Publish(item.ID(), s.pubTopic, new)
}

func (s *SmartBulb) onMessage(client mqtt.Client, msg mqtt.Message) {
	switch msg.Topic() {
	case "smart-bulb1/tick":
		si := s.tick
		old, _ := si.SetValue(ON)
		si.notifyListeners(old, ON)
	}
}

func (s *SmartBulb) TickItem() Item {
	return s.tick
}

func (s *SmartBulb) ColorItem() Item {
	return s.color
}

// NewSmartBulb creates a new SmartBulb Object, publishing and subscribing to the given broker/topic
func NewSmartBulb(id string, label string, conn *MQTTConn, pubTopic string, subTopic string) *SmartBulb {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	s := &SmartBulb{
		id:       id,
		conn:     conn,
		pubTopic: pubTopic,
		subTopic: subTopic,
		tick: &AnItem{
			id:    id + "_TICK",
			label: "Ticker",
			kind:  "value",
		},
		color: &color{
			AnItem: AnItem{
				id:    id + "_COLOR",
				label: "Color",
				kind:  "value",
			},
		},
	}

	conn.Subscribe(subTopic, s)

	s.color.AddListener(s)

	registry.Add(s.tick)
	registry.Add(s.color)

	return s
}
