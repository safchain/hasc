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

package owm

import (
	"fmt"
	"time"

	owm "github.com/briandowns/openweathermap"

	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type OWM struct {
	TemperatureItem *item.AnItem
	HumidityItem    *item.AnItem

	cwd *owm.CurrentWeatherData
	lat float64
	lon float64
}

func (o *OWM) refreshFnc() {
	server.Log.Infof("Weather refresh: %f, %f", o.lat, o.lon)
	o.cwd.CurrentByCoordinates(&owm.Coordinates{Latitude: o.lat, Longitude: o.lon})

	o.TemperatureItem.SetValue(fmt.Sprintf("%.2f", o.cwd.Main.Temp))
	o.HumidityItem.SetValue(fmt.Sprintf("%d", o.cwd.Main.Humidity))
}

func (o *OWM) refresh(refresh time.Duration) {
	// sleep a bit to make the rest of the code ready to accept update
	time.Sleep(5 * time.Second)

	o.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		o.refreshFnc()
	}
}

func NewOWM(id string, label string, apiKey string, lat float64, lon float64, refresh time.Duration) *OWM {
	cwd, err := owm.NewCurrent("C", "EN", apiKey)
	if err != nil {
		server.Log.Fatal(err)
	}

	o := &OWM{
		cwd: cwd,
		lat: lat,
		lon: lon,
		TemperatureItem: &item.AnItem{
			ID:    "TEMPERATURE",
			Label: "Temperature",
			Img:   "temperature",
			Type:  "value",
			Unit:  "°",
		},
		HumidityItem: &item.AnItem{
			ID:    "HUMIDITY",
			Label: "Humidity",
			Img:   "humidity",
			Type:  "value",
			Unit:  "%",
		},
	}

	go o.refresh(refresh)

	server.Registry.Add(o.TemperatureItem, id)
	server.Registry.Add(o.HumidityItem, id)

	return o
}
