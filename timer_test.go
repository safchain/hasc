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
	"testing"
	"time"
)

func TestTimer1(t *testing.T) {
	o := &AnObject{id: "111"}

	tm := newTimer("AAA", "AAA", nil, o, TimerOpts{OnAfter: 1 * time.Second, OffAfter: 2 * time.Second, Timeout: 500 * time.Millisecond})

	if o.State() == ON {
		t.Fatalf("should get OFF state, got: %s", o.State())
	}

	tm.SetState(ON)
	if o.State() == ON {
		t.Fatalf("should get OFF state, got: %s", o.State())
	}
	time.Sleep(600 * time.Millisecond)
	tm.SetState(ON)
	if o.State() == ON {
		t.Fatalf("should get OFF state, got: %s", o.State())
	}

	time.Sleep(300 * time.Millisecond)
	tm.SetState(ON)
	if o.State() == ON {
		t.Fatalf("should get OFF state, got: %s", o.State())
	}
	time.Sleep(300 * time.Millisecond)
	tm.SetState(ON)
	if o.State() != ON {
		t.Fatalf("should get ON state, got: %s", o.State())
	}

	time.Sleep(3 * time.Second)
	if o.State() == ON {
		t.Fatalf("should get OFF state, got: %s", o.State())
	}
}
