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

package envoy

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tidwall/gjson"

	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type Envoy struct {
	TotalProductionItem  *item.AnItem
	TotalConsumptionItem *item.AnItem
	NetConsumptionItem   *item.AnItem
	InvertersItem        *item.AnItem

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
	e.InvertersItem.SetValue(value.String())

	value = gjson.Get(string(body), `production.#(type=="eim").wNow`)
	e.TotalProductionItem.SetValue(value.String())

	value = gjson.Get(string(body), `consumption.#(measurementType=="total-consumption").wNow`)
	e.TotalConsumptionItem.SetValue(value.String())

	value = gjson.Get(string(body), `consumption.#(measurementType=="net-consumption").wNow`)
	e.NetConsumptionItem.SetValue(value.String())
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

func NewEnvoy(id string, label string, endpoint string, refresh time.Duration) *Envoy {
	e := &Envoy{
		endpoint: endpoint,
		TotalProductionItem: &item.AnItem{
			ID:    "TOTAL_PRODUCTION",
			Label: "Total production",
			Img:   "electricity",
			Type:  "value",
			Unit:  "W",
		},
		NetConsumptionItem: &item.AnItem{
			ID:    "NET_CONSUMPTION",
			Label: "Net consumption",
			Img:   "electricity",
			Type:  "value",
			Unit:  "W",
		},
		TotalConsumptionItem: &item.AnItem{
			ID:    "TOTAL_CONSUMPTION",
			Label: "Current",
			Img:   "electricity",
			Type:  "value",
			Unit:  "W",
		},
		InvertersItem: &item.AnItem{
			ID:    "INVERTERS",
			Label: "Inverters",
			Img:   "electricity",
			Type:  "value",
			Unit:  "",
		},
	}

	server.Registry.Add(e.TotalConsumptionItem, id)
	server.Registry.Add(e.TotalProductionItem, id)
	server.Registry.Add(e.NetConsumptionItem, id)
	server.Registry.Add(e.InvertersItem, id)

	go e.refresh(refresh)

	return e
}
