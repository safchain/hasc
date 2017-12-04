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
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT Object
type MQTT struct {
	AnObject
	broker   string
	pubTopic string
	subTopic string
	client   mqtt.Client
}

type MQTTItem struct {
	AnItem
}

func (m *MQTT) onMessage(client mqtt.Client, msg mqtt.Message) {
	m.Lock()
	old := m.state
	new := string(msg.Payload())
	m.state = new
	m.Unlock()
	Log.Infof("MQTT %s changed to %s", m.ID(), new)

	m.notifyListeners(old, new)
}

// SetState changes the internal state of the Object. The new state will be
// published.
func (m *MQTT) SetState(new string) {
	Log.Infof("MQTT %s set to %s", m.ID(), new)

	m.Lock()
	m.state = new
	m.Unlock()

	Log.Infof("MQTT %s send payload: %s", m.ID(), new)
	if token := m.client.Publish(m.pubTopic, 0, false, []byte(new)); token.Wait() && token.Error() != nil {
		Log.Errorf("MQTT error while publishing: %s", token.Error())
	}
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
func newMQTT(id string, label string, broker string, pubTopic string, subTopic string) *MQTT {
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
		broker:   broker,
		pubTopic: pubTopic,
		subTopic: subTopic,
	}

	m.items[ItemID] = &MQTTItem{
		AnItem: AnItem{
			object: m,
			img:    "chart",
		},
	}

	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		Log.Errorf("MQTT connection lost: %s", err)
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		Log.Infof("MQTT connected")

		if token := client.Subscribe(subTopic, 0, m.onMessage); token.Wait() && token.Error() != nil {
			client.Disconnect(0)
			Log.Errorf("MQTT subscription error: %s", token.Error())
		}
	})

	m.client = mqtt.NewClient(opts)

	go func() {
		for {
			if token := m.client.Connect(); token.Wait() && token.Error() != nil {
				Log.Errorf("MQTT connection error: %s", token.Error())
				time.Sleep(2 * time.Second)
			} else {
				return
			}
		}
	}()

	return m
}

// RegisterMQTT registers a MQTT broker. It will publish its state changes to the pubTopic and
// listens for state change on the subTopic.
func RegisterMQTT(id string, label string, broker string, pubTopic string, subTopic string) *MQTT {
	m := newMQTT(id, label, broker, pubTopic, subTopic)
	RegisterObject(m)
	return m
}
