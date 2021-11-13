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
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTT struct {
	AnItem
	pubTopic string
	conn     *MQTTConn
}

func (m *MQTT) onMessage(client mqtt.Client, msg mqtt.Message) {
	new := string(msg.Payload())
	old, _ := m.AnItem.SetValue(new)
	m.notifyListeners(old, new)
}

func (m *MQTT) SetValue(new string) (string, bool) {
	old, updated := m.AnItem.SetValue(new)
	m.conn.Publish(m.ID(), m.pubTopic, new)
	return old, updated
}

func NewMQTT(id string, label string, conn *MQTTConn, pubTopic string, subTopic string) *MQTT {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	m := &MQTT{
		conn:     conn,
		pubTopic: pubTopic,
		AnItem: AnItem{
			id:    id,
			label: label,
			kind:  "value",
			img:   "chart",
		},
	}

	conn.Subscribe(subTopic, m)

	registry.Add(m)

	return m
}
