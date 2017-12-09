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

import "testing"

func TestDecode1(t *testing.T) {
	var msg Message
	if err := msg.Decode("B40192ED1"); err != nil {
		t.Fatal(err)
	}
	if msg.Src != B {
		t.Fatal("Wrong Source")
	}
	if msg.Type != ReadAck {
		t.Fatal("Bad type")
	}
	if len(msg.Values) != 1 {
		t.Fatal("Wrong number of values")
	}
	if msg.Values[0].(float64) != 46.82 {
		t.Fatal("Wrong value")
	}
}

func TestDecode2(t *testing.T) {
	var msg Message
	if err := msg.Decode("BD0010A00"); err != nil {
		t.Fatal(err)
	}
	if msg.Type != WriteAck {
		t.Fatal("Bad type")
	}
	if msg.Src != B {
		t.Fatal("Wrong Source")
	}
	if len(msg.Values) != 1 {
		t.Fatal("Wrong number of values")
	}
	if msg.Values[0].(float64) != 10 {
		t.Fatal("Wrong value")
	}
}

func TestDecode3(t *testing.T) {
	var msg Message
	if err := msg.Decode("T80000200"); err != nil {
		t.Fatal(err)
	}
	if msg.Src != T {
		t.Fatal("Wrong Source")
	}
	if msg.Type != ReadData {
		t.Fatal("Bad type")
	}
	if len(msg.Values) != 2 {
		t.Fatal("Wrong number of values")
	}

	if !msg.IsMasterDHW() || msg.IsMasterCH() {
		t.Fatal("Wrong flag")
	}
}

func TestDecode4(t *testing.T) {
	var msg Message
	if err := msg.Decode("B4000020A"); err != nil {
		t.Fatal(err)
	}
	if msg.Src != B {
		t.Fatal("Wrong Source")
	}
	if msg.Type != ReadAck {
		t.Fatal("Bad type")
	}
	if len(msg.Values) != 2 {
		t.Fatal("Wrong number of values")
	}

	if !msg.IsSlaveCH() || !msg.IsSlaveFlame() || msg.IsSlaveFault() {
		t.Fatal("Wrong flag")
	}
}

func TestEncode(t *testing.T) {
	expected := "B40192ED1"

	var msg Message
	if err := msg.Decode(expected); err != nil {
		t.Fatal(err)
	}

	m, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if m != expected {
		t.Fatalf("got %s expected %s", m, expected)
	}
}
