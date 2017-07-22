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

import "github.com/safchain/hasc"

var (
	device   string
	baud     int
	broker   string
	pubTopic string
	subTopic string
)

func OnStateChange(object hasc.Object, old string, new string) {
	switch object.ID() {
	case "MQTT":
		hasc.ObjectFromID("SERIAL").SetState(new)
	case "SERIAL":
		hasc.ObjectFromID("MQTT").SetState(new)
	}
}

func main() {
	hasc.Cmd.PersistentFlags().StringVarP(&device, "device", "", "/dev/arduino", "serial device, ex: /dev/arduino")
	hasc.Cmd.PersistentFlags().IntVarP(&baud, "baud", "", 9600, "baud, ex: 9600")
	hasc.Cmd.PersistentFlags().StringVarP(&broker, "address", "", "localhost:1883", "MQTT broker address, ex: localhost:1883")
	hasc.Cmd.PersistentFlags().StringVarP(&pubTopic, "pub-topic", "", "/serial-gw/1", "MQTT publisher topic, ex /serial-gw/1")
	hasc.Cmd.PersistentFlags().StringVarP(&subTopic, "sub-topic", "", "/serial-gw/2", "MQTT subscriber topic, ex /serial-gw/2")

	hasc.Start("mqtt-serial-gw", func() {
		hasc.RegisterMQTT("MQTT", "mqtt to serial", "tcp://"+broker, pubTopic, subTopic)
		hasc.RegisterSerial("SERIAL", "serial to mqtt", device, baud)

		// gateway
		hasc.SetStateListener(OnStateChange)
	})
}
