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

package group

import (
	"testing"

	"github.com/safchain/hasc/pkg/item"
)

func TestGroup1(t *testing.T) {
	g := NewGroupItem("AAA", "AAA")

	i1 := &item.AnItem{ID: "111"}
	i2 := &item.AnItem{ID: "222"}

	g.Add(i1)
	g.Add(i2)

	if g.GetValue() != item.OFF {
		t.Fatalf("should get OFF state, got: %s", g.GetValue())
	}

	i1.SetValue(item.ON)
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}

	i1.SetValue(item.OFF)
	g.refresh()
	if g.GetValue() != item.OFF {
		t.Fatalf("should get OFF state, got: %s", g.GetValue())
	}

	i2.SetValue(item.ON)
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}

	i1.SetValue(item.ON)
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}

	i2.SetValue(item.OFF)
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}
}

func TestGroup2(t *testing.T) {
	g := NewGroupItem("AAA", "AAA")

	i1 := &item.AnItem{ID: "111"}
	i2 := &item.AnItem{ID: "222"}

	g.Add(i1)
	g.Add(i2)

	if g.GetValue() != item.OFF {
		t.Fatalf("should get OFF state, got: %s", g.GetValue())
	}

	i1.SetValue(item.ON)
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}

	i1.SetValue("AZE")
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}

	i1.SetValue(item.ON)
	g.refresh()
	if g.GetValue() != item.ON {
		t.Fatalf("should get ON state, got: %s", g.GetValue())
	}
}
