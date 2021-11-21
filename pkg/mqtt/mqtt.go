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

package mqtt

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTListener interface {
	OnValueChange(value string)
}

type MQTT struct {
	ID string

	pubTopic  string
	conn      *MQTTConn
	listeners []MQTTListener
}

func (m *MQTT) AddListener(l MQTTListener) {
	for _, el := range m.listeners {
		if el == l {
			return
		}
	}
	m.listeners = append(m.listeners, l)
}

func (m *MQTT) notifyListeners(value string) {
	for _, l := range m.listeners {
		l.OnValueChange(value)
	}
}

func (m *MQTT) OnMessage(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())
	m.notifyListeners(value)
}

func (m *MQTT) PublishValue(value string) {
	m.conn.Publish(m.ID, m.pubTopic, value)
}

func NewMQTT(id string, conn *MQTTConn, pubTopic string, subTopic string) *MQTT {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	m := &MQTT{
		ID:       id,
		conn:     conn,
		pubTopic: pubTopic,
	}

	conn.Subscribe(subTopic, m)

	return m
}
