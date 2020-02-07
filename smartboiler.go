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
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SmartBoiler Object
type SmartBoiler struct {
	id               string
	pubTopic         string
	subTopic         string
	conn             *MQTTConn
	temperature      *AnItem
	current          *AnItem
	instantFlowMeter *AnItem
	sessionFlowMeter *AnItem
	sessionFlowPrice *AnItem
	relayState       *AnItem
	relayMode        Item
	forceRelayState  Item
}

type force struct {
	AnItem
}

func (f *force) SetValue(new string) (string, bool) {
	old, updated := f.AnItem.SetValue(new)
	f.notifyListeners(old, new)
	return old, updated
}

func (s *SmartBoiler) OnValueChange(item Item, old string, new string) {
	payload := "on"
	if new != ON {
		payload = "off"
	}
	s.conn.Publish(item.ID(), "smab-br/relay", payload)
}

func (s *SmartBoiler) onMessage(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())

	switch msg.Topic() {
	case "smab-br/temperature":
		si := s.temperature
		if si.value != value {
			if old, updated := si.SetValue(value); updated {
				si.notifyListeners(old, value)
			}
		}
	case "smab-br/flow-meter":
		si := s.instantFlowMeter
		newFloat, _ := strconv.ParseFloat(value, 64)
		newFloat = newFloat * 0.5 / 60222 // to liter

		old, _ := si.SetValue(fmt.Sprintf("%.2f", newFloat))
		si.notifyListeners(old, si.value)

		si = s.sessionFlowMeter

		now := time.Now()
		if si.lastValueUpdate.Add(10 * time.Minute).After(now) {
			oldFloat, _ := strconv.ParseFloat(si.value, 64)
			newFloat += oldFloat

			old, _ = si.SetValue(fmt.Sprintf("%.2f", newFloat))
			si.notifyListeners(old, si.value)

			si = s.sessionFlowPrice
			priceFloat := newFloat * 0.3 // 0.3 € / liter

			old, _ = si.SetValue(fmt.Sprintf("%.2f", priceFloat))
			si.notifyListeners(old, si.value)
		} else {
			old, _ := si.SetValue("0.00")
			si.notifyListeners(old, "0.00")

			old, _ = s.SessionPriceItem().SetValue("0.00")
			si.notifyListeners(old, "0.00")
		}
	case "smab-br/current":
		si := s.current
		amp, _ := strconv.ParseFloat(value, 64)
		watt := amp * 220
		old, _ := s.current.SetValue(fmt.Sprintf("%.2f", watt))
		si.notifyListeners(old, value)
	case "smab-br/relay-state":
		si := s.relayState
		if value == "off" {
			if _, updated := si.SetValue(OFF); updated {
				si.notifyListeners(ON, OFF)
			}
		} else if value == "on" {
			if _, updated := si.SetValue(ON); updated {
				si.notifyListeners(OFF, ON)
			}
		}
	}
}

func (s *SmartBoiler) TemperatureItem() Item {
	return s.temperature
}
func (s *SmartBoiler) CurrentSensorItem() Item {
	return s.current
}

func (s *SmartBoiler) InstantFlowMeterItem() Item {
	return s.instantFlowMeter
}

func (s *SmartBoiler) SessionFlowMeterItem() Item {
	return s.sessionFlowMeter
}

func (s *SmartBoiler) SessionPriceItem() Item {
	return s.sessionFlowPrice
}

func (s *SmartBoiler) RelayStateItem() Item {
	return s.relayState
}

func (s *SmartBoiler) RelayModeItem() Item {
	return s.relayMode
}

func (s *SmartBoiler) ForceRelayStateItem() Item {
	return s.forceRelayState
}

// NewSmartBoiler creates a new SmartBoiler Object, publishing and subscribing to the given broker/topic
func NewSmartBoiler(id string, label string, conn *MQTTConn, pubTopic string, subTopic string) *SmartBoiler {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	s := &SmartBoiler{
		id:       id,
		conn:     conn,
		pubTopic: pubTopic,
		subTopic: subTopic,
		temperature: &AnItem{
			id:    id + "_TEMPERATURE",
			label: "Temperature",
			kind:  "value",
			img:   "temperature",
			unit:  "°",
		},
		current: &AnItem{
			id:    id + "_CURRENT",
			label: "Current",
			kind:  "value",
			img:   "electricity",
			unit:  "W",
		},
		instantFlowMeter: &AnItem{
			id:    id + "_INSTANT_METER",
			label: "Instant Liter",
			kind:  "value",
			img:   "shower",
			unit:  "L",
		},
		sessionFlowMeter: &AnItem{
			id:    id + "_SESSION_METER",
			label: "Session Liter",
			kind:  "value",
			img:   "shower",
			unit:  "L",
		},
		sessionFlowPrice: &AnItem{
			id:    id + "_SESSION_PRICE",
			label: "Session Price",
			kind:  "value",
			img:   "price",
			unit:  "€",
		},
		relayState: &AnItem{
			id:    id + "_RELAY_STATE",
			label: "State",
			kind:  "state",
			img:   "plug",
		},
		relayMode: &Switch{
			AnItem: AnItem{
				id:    id + "_RELAY_MODE",
				label: "Mode",
				kind:  "switch",
				img:   "plug",
				value: ON,
			},
		},
		forceRelayState: &force{
			AnItem: AnItem{
				id: id + "_RELAY_FORCE_STATE",
			},
		},
	}

	conn.Subscribe(subTopic, s)

	registry.Add(s.temperature)
	registry.Add(s.current)
	registry.Add(s.instantFlowMeter)
	registry.Add(s.sessionFlowMeter)
	registry.Add(s.sessionFlowPrice)
	registry.Add(s.relayState)
	registry.Add(s.relayMode)
	registry.Add(s.forceRelayState)

	s.relayMode.AddListener(s)
	s.forceRelayState.AddListener(s)

	return s
}
