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

func TestGroup1(t *testing.T) {
	g := newGroup("AAA", "AAA")

	o1 := &AnObject{id: "111"}
	o2 := &AnObject{id: "222"}

	g.Add(o1)
	g.Add(o2)

	if g.State() != OFF {
		t.Fatalf("should get OFF state, got: %s", g.State())
	}

	o1.SetState(ON)
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}

	o1.SetState(OFF)
	g.refresh()
	if g.State() != OFF {
		t.Fatalf("should get OFF state, got: %s", g.State())
	}

	o2.SetState(ON)
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}

	o1.SetState(ON)
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}

	o2.SetState(OFF)
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}
}

func TestGroup2(t *testing.T) {
	g := newGroup("AAA", "AAA")

	o1 := &AnObject{id: "111"}
	o2 := &AnObject{id: "222"}

	g.Add(o1)
	g.Add(o2)

	if g.State() != OFF {
		t.Fatalf("should get OFF state, got: %s", g.State())
	}

	o1.SetState(ON)
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}

	o1.SetState("AZE")
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}

	o1.SetState(ON)
	g.refresh()
	if g.State() != ON {
		t.Fatalf("should get ON state, got: %s", g.State())
	}
}