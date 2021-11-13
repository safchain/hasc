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
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

type Envoy struct {
	TotalProduction  *AnItem
	TotalConsumption *AnItem
	NetConsumption   *AnItem
	Inverters        *AnItem

	endpoint string
}

func (e *Envoy) refreshFnc() {
	resp, err := http.Get(e.endpoint)
	if err != nil {
		log.Panic(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
		return
	}

	value := gjson.Get(string(body), `production.#(type=="inverters").activeCount`)
	old, _ := e.Inverters.SetValue(value.String())
	e.Inverters.notifyListeners(old, value.String())

	value = gjson.Get(string(body), `production.#(type=="eim").wNow`)
	old, _ = e.TotalProduction.SetValue(value.String())
	e.TotalProduction.notifyListeners(old, value.String())

	value = gjson.Get(string(body), `consumption.#(measurementType=="total-consumption").wNow`)
	old, _ = e.TotalConsumption.SetValue(value.String())
	e.TotalConsumption.notifyListeners(old, value.String())

	value = gjson.Get(string(body), `consumption.#(measurementType=="net-consumption").wNow`)
	old, _ = e.NetConsumption.SetValue(value.String())
	e.NetConsumption.notifyListeners(old, value.String())
}

func (e *Envoy) refresh(refresh time.Duration) {
	// sleep a bit to make the rest of the code ready to accept update
	time.Sleep(5 * time.Second)

	e.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		e.refreshFnc()
	}
}

func (e *Envoy) TotalProductionItem() Item {
	return e.TotalProduction
}

func (e *Envoy) TotalConsumptionItem() Item {
	return e.TotalConsumption
}

func (e *Envoy) NetConsumptionItem() Item {
	return e.NetConsumption
}

func (e *Envoy) InvertersItem() Item {
	return e.Inverters
}

func NewEnvoy(id string, label string, endpoint string, refresh time.Duration) *Envoy {
	e := &Envoy{
		endpoint: endpoint,
		TotalProduction: &AnItem{
			id:    id + "_TOTAL_PRODUCTION",
			label: "Total production",
			img:   "electricity",
			kind:  "value",
			unit:  "W",
		},
		NetConsumption: &AnItem{
			id:    id + "_NET_CONSUMPTION",
			label: "Net consumption",
			img:   "electricity",
			kind:  "value",
			unit:  "W",
		},
		TotalConsumption: &AnItem{
			id:    id + "_TOTAL_CONSUMPTION",
			label: "Current",
			img:   "electricity",
			kind:  "value",
			unit:  "W",
		},
		Inverters: &AnItem{
			id:    id + "_INVERTERS",
			label: "Inverters",
			img:   "electricity",
			kind:  "value",
			unit:  "",
		},
	}

	registry.Add(e.TotalConsumption)
	registry.Add(e.TotalProduction)
	registry.Add(e.NetConsumption)
	registry.Add(e.Inverters)

	go e.refresh(refresh)

	return e
}
