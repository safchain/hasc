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

package smartbulb

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/safchain/hasc/pkg/item"
	hmqtt "github.com/safchain/hasc/pkg/mqtt"
	"github.com/safchain/hasc/pkg/server"
)

// SmartBulb Object
type SmartBulb struct {
	TickItem  *item.AnItem
	ColorItem item.Item

	id       string
	pubTopic string
	subTopic string
	conn     *hmqtt.MQTTConn
}

type color struct {
	item.AnItem
}

func (s *SmartBulb) OnValueChange(it item.Item, old string, new string) {
	s.conn.Publish(it.GetID(), s.pubTopic, new)
}

func (s *SmartBulb) OnMessage(client mqtt.Client, msg mqtt.Message) {
	switch msg.Topic() {
	case "smart-bulb1/tick":
		s.TickItem.SetValue(item.ON)
	}
}

// NewSmartBulb creates a new SmartBulb Object, publishing and subscribing to the given broker/topic
func NewSmartBulb(id string, label string, conn *hmqtt.MQTTConn, pubTopic string, subTopic string) *SmartBulb {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	s := &SmartBulb{
		id:       id,
		conn:     conn,
		pubTopic: pubTopic,
		subTopic: subTopic,
		TickItem: &item.AnItem{
			ID:    "TICK",
			Label: "Ticker",
			Type:  "value",
		},
		ColorItem: &color{
			AnItem: item.AnItem{
				ID:    "COLOR",
				Label: "Color",
				Type:  "value",
			},
		},
	}

	conn.Subscribe(subTopic, s)

	s.ColorItem.AddListener(s)

	server.Registry.Add(s.TickItem, id)
	server.Registry.Add(s.ColorItem, id)

	return s
}
