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
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"

	owm "github.com/briandowns/openweathermap"
)

type owmValues struct {
	Temperature float64
	Humidity    int
}

type OWM struct {
	AnObject
	cwd    *owm.CurrentWeatherData
	lat    float64
	lon    float64
	values owmValues
}

type TemperatureItem struct {
	AnItem
}

type HummidityItem struct {
	AnItem
}

const (
	TemperatureID = "TEMPERATURE"
	HummidityID   = "HUMIDITY"
)

func (o *OWM) setState(new string) {
	Log.Infof("Weather %s set to %s", o.ID(), new)

	o.Lock()
	old := o.state
	o.state = new
	o.Unlock()

	if old != new {
		o.notifyListeners(old, new)
	}
}

func (o *OWM) refreshFnc() {
	Log.Infof("Weather %s refresh", o.ID())
	o.cwd.CurrentByCoordinates(&owm.Coordinates{Latitude: o.lat, Longitude: o.lon})

	o.values.Temperature = o.cwd.Main.Temp
	o.values.Humidity = o.cwd.Main.Humidity

	data, _ := json.Marshal(o.values)
	o.setState(string(data))
}

func (o *OWM) refresh(refresh time.Duration) {
	o.refreshFnc()

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		o.refreshFnc()
	}
}

func (o *OWM) TemperatureItem() Item {
	return o.items[TemperatureID]
}

func (ti *TemperatureItem) ID() string {
	return TemperatureID
}

func (ti *TemperatureItem) Value() string {
	ti.object.RLock()
	defer ti.object.RUnlock()

	return fmt.Sprintf("%.2f", ti.object.(*OWM).values.Temperature)
}

func (ti *TemperatureItem) HTML() template.HTML {
	ti.object.RLock()
	defer ti.object.RUnlock()

	data := struct {
		ID        string
		Label     string
		ItemLabel string
		Unit      string
		Img       string
	}{
		ID:        ti.object.ID() + "_" + ti.ID(),
		Label:     ti.object.Label(),
		ItemLabel: ti.Label(),
		Unit:      "°",
		Img:       fmt.Sprintf("statics/img/%s.png", ti.img),
	}

	return itemTemplate("statics/items/value.html", data)
}

func (ti *TemperatureItem) Label() string {
	return "Temp."
}

func (ti *TemperatureItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(ti)
}

func (o *OWM) HummidityItem() Item {
	return o.items[HummidityID]
}

func (hi *HummidityItem) ID() string {
	return HummidityID
}

func (hi *HummidityItem) Value() string {
	hi.object.RLock()
	defer hi.object.RUnlock()

	return fmt.Sprintf("%d", hi.object.(*OWM).values.Humidity)
}

func (hi *HummidityItem) HTML() template.HTML {
	hi.object.RLock()
	defer hi.object.RUnlock()

	data := struct {
		ID        string
		Label     string
		ItemLabel string
		Unit      string
		Img       string
	}{
		ID:        hi.object.ID() + "_" + hi.ID(),
		Label:     hi.object.Label(),
		ItemLabel: hi.Label(),
		Unit:      "%",
		Img:       fmt.Sprintf("statics/img/%s.png", hi.img),
	}

	return itemTemplate("statics/items/value.html", data)
}

func (hi *HummidityItem) Label() string {
	return "Humidity"
}

func (hi *HummidityItem) MarshalJSON() ([]byte, error) {
	return marshalJSON(hi)
}

func newOWM(id string, label string, apiKey string, lat float64, lon float64, refresh time.Duration) *OWM {
	os.Setenv("OWM_API_KEY", apiKey)

	cwd, err := owm.NewCurrent("C", "EN")
	if err != nil {
		Log.Fatal(err)
	}

	o := &OWM{
		AnObject: AnObject{
			id:    id,
			label: label,
			items: make(map[string]Item),
		},
		cwd: cwd,
		lat: lat,
		lon: lon,
	}

	o.items[TemperatureID] = &TemperatureItem{
		AnItem: AnItem{
			object: o,
			img:    "temperature",
		},
	}

	o.items[HummidityID] = &HummidityItem{
		AnItem: AnItem{
			object: o,
			img:    "humidity",
		},
	}

	go o.refresh(refresh)

	return o
}

// RegisterOWM registers an OpenWeatherMap monitor. It reports temperature and humidity.
func RegisterOWM(id string, label string, apiKey string, lat float64, lon float64, refresh time.Duration) *OWM {
	o := newOWM(id, label, apiKey, lat, lon, refresh)
	RegisterObject(o)
	return o
}
