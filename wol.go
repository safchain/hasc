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

type WOL struct {
	mac  string
	intf string
}

func (w *WOL) OnStateChange(object Object, old string, new string) {
	switch new {
	case "on", "ON", "1":
		// TODO : fix up wol

		/*if err := wol.SendMagicPacket(w.mac, "255.255.255.255:9", w.intf); err != nil {
			Log.Errorf("WOL unable to send magic packet: %s", err)
		}*/
	}
}

// NewWOL creates a new Wake-On-Lan for the given MAC and using
// the given interface name to send the request.
func NewWOL(mac string, intf string) *WOL {
	Log.Infof("New WOL %s %s", mac, intf)

	return &WOL{
		mac:  mac,
		intf: intf,
	}
}
