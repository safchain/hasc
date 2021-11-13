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
	"fmt"
	"time"

	owm "github.com/briandowns/openweathermap"
)

type OWM struct {
	temperature AnItem
	humidity    AnItem
	cwd         *owm.CurrentWeatherData
	lat         float64
	lon         float64
}

func (o *OWM) refreshFnc() {
	Log.Infof("Weather refresh: %f, %f", o.lat, o.lon)
	o.cwd.CurrentByCoordinates(&owm.Coordinates{Latitude: o.lat, Longitude: o.lon})

	old, _ := o.temperature.SetValue(fmt.Sprintf("%.2f", o.cwd.Main.Temp))
	o.temperature.notifyListeners(old, o.temperature.value)

	old, _ = o.humidity.SetValue(fmt.Sprintf("%d", o.cwd.Main.Humidity))
	o.humidity.notifyListeners(old, o.humidity.value)
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

func (o *OWM) TemperatureItem() Item {
	return &o.temperature
}

func (o *OWM) HummidityItem() Item {
	return &o.humidity
}

func NewOWM(id string, label string, apiKey string, lat float64, lon float64, refresh time.Duration) *OWM {
	cwd, err := owm.NewCurrent("C", "EN", apiKey)
	if err != nil {
		Log.Fatal(err)
	}

	o := &OWM{
		cwd: cwd,
		lat: lat,
		lon: lon,
		temperature: AnItem{
			id:    id + "_TEMPERATURE",
			label: "Temperature",
			img:   "temperature",
			kind:  "value",
			unit:  "°",
		},
		humidity: AnItem{
			id:    id + "_HUMIDITY",
			label: "Humidity",
			img:   "humidity",
			kind:  "value",
			unit:  "%",
		},
	}

	go o.refresh(refresh)

	registry.Add(&o.temperature)
	registry.Add(&o.humidity)

	return o
}
