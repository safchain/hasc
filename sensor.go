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

type Sensor struct {
	*Value
}

func (s *Sensor) SetState(new string) {
	Log.Infof("Sensor %s set to %s", s.ID(), new)

	s.Lock()
	old := s.state
	s.state = new
	s.Unlock()

	s.notifyListeners(old, new)
}

func newSensor(id string, label string, device interface{}, unit string) *Sensor {
	s := &Sensor{
		Value: newValue(id, label, unit),
	}

	s.device = device

	return s
}

// RegisterSensor registers a new sensor.
func RegisterSensor(id string, label string, device interface{}, unit string) *Sensor {
	s := newSensor(id, label, device, unit)
	RegisterObject(s)
	return s
}
