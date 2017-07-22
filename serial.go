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
	"strings"
	"time"

	"github.com/tarm/serial"
)

type Serial struct {
	AnObject
	dev  string
	baud int
	port *serial.Port
}

func (s *Serial) read() {
	s.RLock()
	port := s.port
	s.RUnlock()

	buf := make([]byte, 128)
	for {
		n, err := port.Read(buf)
		if err != nil {
			// try to re-open
			if err = s.openPort(); err == nil {
				s.RLock()
				port = s.port
				s.RUnlock()
			}
			time.Sleep(time.Second)
		}
		if n > 0 {
			s.Lock()
			old := s.state
			new := strings.TrimSpace(string(buf[0:n]))
			s.state = new
			s.Unlock()

			Log.Infof("Serial %s changed to %s", s.ID(), new)

			s.notifyListeners(old, new)
		}
	}
}

func (s *Serial) SetState(new string) {
	Log.Infof("Serial %s set to %s", s.ID(), new)

	s.Lock()
	old := s.state
	s.state = new
	s.Unlock()

	s.notifyListeners(old, new)

	Log.Infof("Serial %s send payload: %s", s.ID(), new)

	s.Lock()
	port := s.port
	s.Unlock()

	if _, err := port.Write([]byte(new + "\n")); err != nil {
		s.openPort()
	}
}

func (s *Serial) openPort() error {
	s.Lock()
	defer s.Unlock()

	c := &serial.Config{Name: s.dev, Baud: s.baud}
	port, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	s.port = port

	return nil
}

func newSerial(id string, label string, dev string, baud int) *Serial {
	s := &Serial{
		AnObject: AnObject{
			id:    id,
			label: label,
		},
		dev:  dev,
		baud: baud,
	}

	if err := s.openPort(); err != nil {
		Log.Fatalf("Unable to open serial port %s: %s", dev, err)
	}

	go s.read()

	return s
}

// RegisterSerial listens serial device. Set the state with the value read from the serial port
// and write to the serial port the state changes.
func RegisterSerial(id string, label string, dev string, baud int) *Serial {
	s := newSerial(id, label, dev, baud)
	RegisterObject(s)
	return s
}
