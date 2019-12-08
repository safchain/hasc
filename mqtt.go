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
	"html/template"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT Object
type MQTT struct {
	AnObject
	pubTopic string
	subTopic string
	conn     *MQTTConn
}

type MQTTItem struct {
	AnItem
}

func (m *MQTT) onMessage(client mqtt.Client, msg mqtt.Message) {
	new := string(msg.Payload())
	old := m.AnObject.SetState(new)

	Log.Infof("MQTT %s changed to %s", m.ID(), new)

	m.notifyListeners(old, new)
}

// SetState changes the internal state of the Object. The new state will be
// published.
func (m *MQTT) SetState(new string) string {
	Log.Infof("MQTT %s set to %s", m.ID(), new)

	old := m.AnObject.SetState(new)

	m.conn.Publish(m.ID(), m.pubTopic, new)

	return old
}

func (mi *MQTTItem) HTML() template.HTML {
	mi.object.RLock()
	defer mi.object.RUnlock()

	return valueTemplate(mi, "Value", "", mi.img)
}

func (mi *MQTTItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(mi)
}

// NewMQTT creates a new MQTT Object, publishing and subscribing to the given broker/topic
func newMQTT(id string, label string, conn *MQTTConn, pubTopic string, subTopic string) *MQTT {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	m := &MQTT{
		AnObject: AnObject{
			id:    id,
			label: label,
			items: make(map[string]Item),
		},
		conn:     conn,
		pubTopic: pubTopic,
		subTopic: subTopic,
	}

	m.items[ItemID] = &MQTTItem{
		AnItem: AnItem{
			object: m,
			img:    "chart",
		},
	}

	conn.Subscribe(subTopic, m)

	return m
}

// RegisterMQTT registers a MQTT broker. It will publish its state changes to the pubTopic and
// listens for state change on the subTopic.
func RegisterMQTT(id string, label string, conn *MQTTConn, pubTopic string, subTopic string) *MQTT {
	m := newMQTT(id, label, conn, pubTopic, subTopic)
	RegisterObject(m)
	return m
}
