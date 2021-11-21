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

package serial

import (
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"

	"github.com/safchain/hasc/pkg/server"
)

type SerialListener interface {
	OnValueChange(value string)
}

type CallbackListener struct {
	CbFnc func(value string)
}

func (c *CallbackListener) OnValueChange(value string) {
	c.CbFnc(value)
}

type Serial struct {
	sync.RWMutex

	dev       string
	baud      int
	port      *serial.Port
	listeners []SerialListener
}

func (s *Serial) AddListener(l SerialListener) {
	for _, el := range s.listeners {
		if el == l {
			return
		}
	}
	s.listeners = append(s.listeners, l)
}

func (s *Serial) notifyListeners(value string) {
	for _, l := range s.listeners {
		l.OnValueChange(value)
	}
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
			value := strings.TrimSpace(string(buf[0:n]))
			s.notifyListeners(value)
		}
	}
}

func (s *Serial) WriteValue(new string) {
	if _, err := s.port.Write([]byte(new + "\n")); err != nil {
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

func NewSerial(dev string, baud int) *Serial {
	s := &Serial{
		dev:  dev,
		baud: baud,
	}

	if err := s.openPort(); err != nil {
		server.Log.Fatalf("Unable to open serial port %s: %s", dev, err)
	}

	go s.read()

	return s
}
