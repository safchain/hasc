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

package smartboiler

import (
	"fmt"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/safchain/hasc/pkg/button"
	"github.com/safchain/hasc/pkg/item"
	hmqtt "github.com/safchain/hasc/pkg/mqtt"
	"github.com/safchain/hasc/pkg/server"
)

// SmartBoiler Object
type SmartBoiler struct {
	TemperatureItem      *item.AnItem
	CurrentItem          *item.AnItem
	InstantFlowMeterItem *item.AnItem
	SessionFlowMeterItem *item.AnItem
	SessionFlowPriceItem *item.AnItem
	RelayStateItem       *item.AnItem
	RelayModeItem        item.Item
	ForceRelayStateItem  item.Item

	id       string
	pubTopic string
	subTopic string
	conn     *hmqtt.MQTTConn
}

type force struct {
	item.AnItem
}

func (s *SmartBoiler) OnValueChange(it item.Item, old string, new string) {
	payload := "on"
	if new != item.ON {
		payload = "off"
	}
	s.conn.Publish(it.GetID(), "smab-br/relay", payload)
}

func (s *SmartBoiler) OnMessage(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())

	switch msg.Topic() {
	case "smab-br/temperature":
		s.TemperatureItem.SetValue(value)
	case "smab-br/flow-meter":
		now := time.Now()

		// InstantFlowMeterItem
		si := s.InstantFlowMeterItem
		value, _ := strconv.ParseFloat(value, 64)

		newFloat := value * 0.5 / 34887                                           // to liter
		literPerMin := newFloat / now.Sub(si.GetLastValueChange()).Seconds() * 60 // per min

		si.SetValue(fmt.Sprintf("%.4f", literPerMin))

		// SessionFlowMeterItem
		si = s.SessionFlowMeterItem
		oldFloat, _ := strconv.ParseFloat(si.GetValue(), 64)

		if si.GetLastValueChange().Add(2 * time.Minute).Before(now) {
			oldFloat = 0
		}
		newFloat += oldFloat

		si.SetValue(fmt.Sprintf("%.6f", newFloat))

		// SessionFlowPriceItem
		si = s.SessionFlowPriceItem
		priceFloat := newFloat * 3 / 1000 // 3 € / m3

		si.SetValue(fmt.Sprintf("%.5f", priceFloat))
	case "smab-br/current":
		amp, _ := strconv.ParseFloat(value, 64)
		watt := amp * 220
		if watt < 500 {
			watt = 0
		}

		s.CurrentItem.SetValue(fmt.Sprintf("%.2f", watt))
	case "smab-br/relay-state":
		si := s.RelayStateItem
		if value == "off" {
			si.SetValue(item.OFF)
		} else if value == "on" {
			si.SetValue(item.ON)
		}
	}
}

// NewSmartBoiler creates a new SmartBoiler Object, publishing and subscribing to the given broker/topic
func NewSmartBoiler(id string, label string, conn *hmqtt.MQTTConn, pubTopic string, subTopic string) *SmartBoiler {
	if pubTopic == subTopic {
		fmt.Println("pub topic and sub topic have to be different")
		os.Exit(1)
	}

	s := &SmartBoiler{
		id:       id,
		conn:     conn,
		pubTopic: pubTopic,
		subTopic: subTopic,
		TemperatureItem: &item.AnItem{
			ID:    fmt.Sprintf("%s/TEMPERATURE", id),
			Label: "Temperature",
			Type:  "value",
			Img:   "temperature",
			Unit:  "°",
		},
		CurrentItem: &item.AnItem{
			ID:    fmt.Sprintf("%s/CURRENT", id),
			Label: "Current",
			Type:  "value",
			Img:   "electricity",
			Unit:  "W",
		},
		InstantFlowMeterItem: &item.AnItem{
			ID:    fmt.Sprintf("%s/INSTANT_METER", id),
			Label: "Instant Liter",
			Type:  "value",
			Img:   "shower",
			Unit:  "L/M",
		},
		SessionFlowMeterItem: &item.AnItem{
			ID:    fmt.Sprintf("%s/SESSION_METER", id),
			Label: "Session Liter",
			Type:  "value",
			Img:   "shower",
			Unit:  "L",
		},
		SessionFlowPriceItem: &item.AnItem{
			ID:    fmt.Sprintf("%s/SESSION_PRICE", id),
			Label: "Session Price",
			Type:  "value",
			Img:   "price",
			Unit:  "€",
		},
		RelayStateItem: &item.AnItem{
			ID:    fmt.Sprintf("%s/RELAY_STATE", id),
			Label: "State",
			Type:  "state",
			Img:   "plug",
		},
		RelayModeItem: &button.SwitchItem{
			AnItem: item.AnItem{
				ID:    fmt.Sprintf("%s/RELAY_MODE", id),
				Label: "Mode",
				Type:  "switch",
				Img:   "plug",
			},
		},
		ForceRelayStateItem: &force{
			AnItem: item.AnItem{
				ID: "RELAY_FORCE_STATE",
			},
		},
	}

	s.InstantFlowMeterItem.SetValue("0")
	s.InstantFlowMeterItem.SetValue("0")
	s.RelayModeItem.SetValue(item.ON)

	conn.Subscribe(subTopic, s)

	server.Registry.Add(s.TemperatureItem)
	server.Registry.Add(s.CurrentItem)
	server.Registry.Add(s.InstantFlowMeterItem)
	server.Registry.Add(s.SessionFlowMeterItem)
	server.Registry.Add(s.SessionFlowPriceItem)
	server.Registry.Add(s.RelayStateItem)
	server.Registry.Add(s.RelayModeItem)
	server.Registry.Add(s.ForceRelayStateItem)

	s.RelayModeItem.AddListener(s)
	s.ForceRelayStateItem.AddListener(s)

	return s
}
