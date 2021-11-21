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

package main

import (
	"github.com/spf13/cobra"

	"github.com/safchain/hasc/pkg/mqtt"
	"github.com/safchain/hasc/pkg/serial"
)

var (
	cmd      *cobra.Command
	device   string
	baud     int
	broker   string
	pubTopic string
	subTopic string
)

type serialListener struct {
	mqtt *mqtt.MQTT
}

type mqttListener struct {
	serial *serial.Serial
}

func (s *serialListener) OnValueChange(value string) {
	s.mqtt.PublishValue(value)
}

func (m *mqttListener) OnValueChange(value string) {
	m.serial.WriteValue(value)
}

func main() {
	var cmd cobra.Command

	cmd.PersistentFlags().StringVarP(&device, "device", "", "/dev/arduino", "serial device, ex: /dev/arduino")
	cmd.PersistentFlags().IntVarP(&baud, "baud", "", 9600, "baud, ex: 9600")
	cmd.PersistentFlags().StringVarP(&broker, "address", "", "localhost:1883", "MQTT broker address, ex: localhost:1883")
	cmd.PersistentFlags().StringVarP(&pubTopic, "pub-topic", "", "/serial-gw/1", "MQTT publisher topic, ex /serial-gw/1")
	cmd.PersistentFlags().StringVarP(&subTopic, "sub-topic", "", "/serial-gw/2", "MQTT subscriber topic, ex /serial-gw/2")

	conn := mqtt.NewMQTTConn("tcp://" + broker)
	m := mqtt.NewMQTT("MQTT", conn, pubTopic, subTopic)
	s := serial.NewSerial(device, baud)

	ml := &mqttListener{serial: s}
	sl := &serialListener{mqtt: m}

	m.AddListener(ml)
	s.AddListener(sl)

	ch := make(chan bool)
	<-ch
}

func init() {

}
