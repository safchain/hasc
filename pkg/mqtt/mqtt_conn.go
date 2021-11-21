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
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/safchain/hasc/pkg/server"
)

type MessageHandler interface {
	OnMessage(client mqtt.Client, msg mqtt.Message)
}

type Subscriber struct {
	topic   string
	handler MessageHandler
}

// MQTTConn Object
type MQTTConn struct {
	sync.RWMutex
	connected   int64
	broker      string
	client      mqtt.Client
	subscribers []*Subscriber
}

func (m *MQTTConn) Subscribe(topic string, handler MessageHandler) {
	subscriber := &Subscriber{
		topic:   topic,
		handler: handler,
	}

	m.Lock()
	m.subscribers = append(m.subscribers, subscriber)
	m.Unlock()

	if atomic.LoadInt64(&m.connected) == 1 {
		if err := m.subscribe(subscriber); err != nil {
			server.Log.Errorf("MQTT subscription error: %s", err)
		}
	}
}

func (m *MQTTConn) subscribe(subscriber *Subscriber) error {
	server.Log.Infof("MQTT subscribe to: %s", subscriber.topic)
	if token := m.client.Subscribe(subscriber.topic, 0, subscriber.handler.OnMessage); token.Wait() && token.Error() != nil {
		m.client.Disconnect(0)
		return token.Error()
	}

	return nil
}

func (m *MQTTConn) Publish(id, topic, payload string) {
	server.Log.Infof("MQTT %s send payload: %s", id, payload)
	if token := m.client.Publish(topic, 0, false, []byte(payload)); token.Wait() && token.Error() != nil {
		server.Log.Errorf("MQTT error while publishing: %s", token.Error())
	}
}

func (m *MQTTConn) subscribeAll() {
	if atomic.LoadInt64(&m.connected) == 1 {
		m.RLock()
		for _, s := range m.subscribers {
			if err := m.subscribe(s); err != nil {
				server.Log.Errorf("MQTT subscription error: %s", err)
			}
		}
	}
}

// NewMQTTConn creates a new MQTTConn Object, publishing and subscribing to the given broker/topic
func NewMQTTConn(broker string) *MQTTConn {
	m := &MQTTConn{
		broker: broker,
	}

	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		server.Log.Errorf("MQTT connection lost: %s", err)

		atomic.StoreInt64(&m.connected, 0)
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		server.Log.Infof("MQTT connected")

		atomic.StoreInt64(&m.connected, 1)
		m.subscribeAll()
	})

	m.client = mqtt.NewClient(opts)

	go func() {
		for {
			if token := m.client.Connect(); token.Wait() && token.Error() != nil {
				server.Log.Errorf("MQTT connection error: %s", token.Error())
				time.Sleep(2 * time.Second)
			} else {
				return
			}
		}
	}()

	return m
}
