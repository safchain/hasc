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

package opentherm

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/safchain/hasc/pkg/button"
	"github.com/safchain/hasc/pkg/item"
	hmqtt "github.com/safchain/hasc/pkg/mqtt"
	"github.com/safchain/hasc/pkg/server"
	"github.com/safchain/hasc/pkg/value"
)

type argType int

const (
	flag8 argType = iota + 1
	f8
	u8
	u16
	nu
	s8
	dt
	tm
	ns
)

type MessageType int

const (
	ReadData MessageType = iota
	WriteData
	InvData
	Reserved
	ReadAck
	WriteAck
	DataInv
	UnkDataId
)

type Flag int

const (
	MasterCH          Flag = 1
	MasterDHW         Flag = 2
	MasterCooling     Flag = 4
	MasterOTC         Flag = 8
	MasterCH2         Flag = 16
	MasterSummer      Flag = 32
	MasterDHWBlocking Flag = 64
	SlaveFault        Flag = 1
	SlaveCH           Flag = 2
	SlaveDHW          Flag = 4
	SlaveFlame        Flag = 8
	SlaveCooling      Flag = 16
	SlaveCH2          Flag = 32
	SlaveDiagnostic   Flag = 64
	SlaveElectricity  Flag = 128
)

type Src rune

const (
	B Src = 'B'
	T Src = 'T'
	A Src = 'A'
	R Src = 'R'
)

type Message struct {
	ID     int
	Src    Src
	Type   MessageType
	Values []interface{}
	Desc   string
}

type messageDef struct {
	id   int
	arg1 argType
	arg2 argType
	desc string
}

var messageDefs = map[int]messageDef{
	0:   {id: 0, arg1: flag8, arg2: flag8, desc: "Status"},
	1:   {id: 1, arg1: f8, arg2: ns, desc: "Control setpoint"},
	5:   {id: 5, arg1: flag8, arg2: u8, desc: "Application-specific flags"},
	8:   {id: 8, arg1: f8, arg2: ns, desc: "Control setpoint 2"},
	70:  {id: 70, arg1: flag8, arg2: flag8, desc: "Status V/H"},
	71:  {id: 71, arg1: nu, arg2: u8, desc: "Control setpoint V/H"},
	72:  {id: 72, arg1: flag8, arg2: u8, desc: "Fault flags/code V/H"},
	73:  {id: 73, arg1: u16, arg2: ns, desc: "OEM diagnostic code V/H"},
	101: {id: 101, arg1: flag8, arg2: flag8, desc: "Solar storage mode and status"},
	102: {id: 102, arg1: flag8, arg2: u8, desc: "Solar storage fault flags"},
	115: {id: 115, arg1: u16, arg2: ns, desc: "OEM diagnostic code"},
	2:   {id: 2, arg1: flag8, arg2: u8, desc: "Master configuration"},
	3:   {id: 3, arg1: flag8, arg2: u8, desc: "Slave configuration"},
	74:  {id: 74, arg1: flag8, arg2: u8, desc: "Configuration/memberid V/H"},
	75:  {id: 75, arg1: f8, arg2: ns, desc: "OpenTherm version V/H"},
	76:  {id: 76, arg1: u8, arg2: u8, desc: "Product version V/H"},
	103: {id: 103, arg1: flag8, arg2: u8, desc: "Solar storage config/memberid"},
	104: {id: 104, arg1: u8, arg2: u8, desc: "Solar storage product version"},
	124: {id: 124, arg1: f8, arg2: ns, desc: "OpenTherm version Master"},
	125: {id: 125, arg1: f8, arg2: ns, desc: "OpenTherm version Slave"},
	126: {id: 126, arg1: u8, arg2: u8, desc: "Master product version"},
	127: {id: 127, arg1: u8, arg2: u8, desc: "Slave product version"},
	4:   {id: 4, arg1: u8, arg2: u8, desc: "Remote command"},
	16:  {id: 16, arg1: f8, arg2: ns, desc: "Room setpoint"},
	17:  {id: 17, arg1: f8, arg2: ns, desc: "Relative modulation level"},
	18:  {id: 18, arg1: f8, arg2: ns, desc: "CH water pressure"},
	19:  {id: 19, arg1: f8, arg2: ns, desc: "DHW flow rate"},
	20:  {id: 20, arg1: tm, arg2: ns, desc: "Day of week and tm of day"},
	21:  {id: 21, arg1: dt, arg2: ns, desc: "Date"},
	22:  {id: 22, arg1: u16, arg2: ns, desc: "Year"},
	23:  {id: 23, arg1: f8, arg2: ns, desc: "Room Setpoint CH2"},
	24:  {id: 24, arg1: f8, arg2: ns, desc: "Room temperature"},
	25:  {id: 25, arg1: f8, arg2: ns, desc: "Boiler water temperature"},
	26:  {id: 26, arg1: f8, arg2: ns, desc: "DHW temperature"},
	27:  {id: 27, arg1: f8, arg2: ns, desc: "Outside temperature"},
	28:  {id: 28, arg1: f8, arg2: ns, desc: "Return water temperature"},
	29:  {id: 29, arg1: f8, arg2: ns, desc: "Solar storage temperature"},
	30:  {id: 30, arg1: f8, arg2: ns, desc: "Solar collector temperature"},
	31:  {id: 31, arg1: f8, arg2: ns, desc: "Flow temperature CH2"},
	32:  {id: 32, arg1: f8, arg2: ns, desc: "DHW2 temperature"},
	33:  {id: 33, arg1: f8, arg2: ns, desc: "Exhaust temperature"},
	34:  {id: 34, arg1: f8, arg2: ns, desc: "Boiler heat exchanger temperature"},
	35:  {id: 35, arg1: u8, arg2: u8, desc: "Boiler fan speed and setpoint"},
	37:  {id: 37, arg1: f8, arg2: ns, desc: "Room temperature CH2"},
	77:  {id: 77, arg1: nu, arg2: u8, desc: "Relative ventilation"},
	78:  {id: 78, arg1: u8, arg2: u8, desc: "Relative humidity exhaust air"},
	79:  {id: 79, arg1: u16, arg2: ns, desc: "CO2 level exhaust air"},
	80:  {id: 80, arg1: f8, arg2: ns, desc: "Supply inlet temperature"},
	81:  {id: 81, arg1: f8, arg2: ns, desc: "Supply outlet temperature"},
	82:  {id: 82, arg1: f8, arg2: ns, desc: "Exhaust inlet temperature"},
	83:  {id: 83, arg1: f8, arg2: ns, desc: "Exhaust outlet temperature"},
	84:  {id: 84, arg1: u16, arg2: ns, desc: "Exhaust fan speed"},
	85:  {id: 85, arg1: u16, arg2: ns, desc: "Inlet fan speed"},
	113: {id: 113, arg1: u16, arg2: ns, desc: "Unsuccessful burner starts"},
	114: {id: 114, arg1: u16, arg2: ns, desc: "Flame signal too low count"},
	116: {id: 116, arg1: u16, arg2: ns, desc: "Burner starts"},
	117: {id: 117, arg1: u16, arg2: ns, desc: "CH pump starts"},
	118: {id: 118, arg1: u16, arg2: ns, desc: "DHW pump/valve starts"},
	119: {id: 119, arg1: u16, arg2: ns, desc: "DHW burner starts"},
	120: {id: 120, arg1: u16, arg2: ns, desc: "Burner operation hours"},
	121: {id: 121, arg1: u16, arg2: ns, desc: "CH pump operation hours"},
	122: {id: 122, arg1: u16, arg2: ns, desc: "DHW pump/valve operation hours"},
	123: {id: 123, arg1: u16, arg2: ns, desc: "DHW burner operation hours"},
	6:   {id: 6, arg1: flag8, arg2: flag8, desc: "Remote parameter flags"},
	48:  {id: 48, arg1: s8, arg2: s8, desc: "DHW setpoint boundaries"},
	49:  {id: 49, arg1: s8, arg2: s8, desc: "Max CH setpoint boundaries"},
	50:  {id: 50, arg1: s8, arg2: s8, desc: "OTC heat curve ratio boundaries"},
	51:  {id: 51, arg1: s8, arg2: s8, desc: "Remote parameter 4 boundaries"},
	52:  {id: 52, arg1: s8, arg2: s8, desc: "Remote parameter 5 boundaries"},
	53:  {id: 53, arg1: s8, arg2: s8, desc: "Remote parameter 6 boundaries"},
	54:  {id: 54, arg1: s8, arg2: s8, desc: "Remote parameter 7 boundaries"},
	55:  {id: 55, arg1: s8, arg2: s8, desc: "Remote parameter 8 boundaries"},
	56:  {id: 56, arg1: f8, arg2: ns, desc: "DHW setpoint"},
	57:  {id: 57, arg1: f8, arg2: ns, desc: "Max CH water setpoint"},
	58:  {id: 58, arg1: f8, arg2: ns, desc: "OTC heat curve ratio"},
	59:  {id: 59, arg1: f8, arg2: ns, desc: "Remote parameter 4"},
	60:  {id: 60, arg1: f8, arg2: ns, desc: "Remote parameter 5"},
	61:  {id: 61, arg1: f8, arg2: ns, desc: "Remote parameter 6"},
	62:  {id: 62, arg1: f8, arg2: ns, desc: "Remote parameter 7"},
	63:  {id: 63, arg1: f8, arg2: ns, desc: "Remote parameter 8"},
	86:  {id: 86, arg1: flag8, arg2: flag8, desc: "Remote parameter settings V/H"},
	87:  {id: 87, arg1: u8, arg2: nu, desc: "Nominal ventilation value"},
	10:  {id: 10, arg1: u8, arg2: nu, desc: "Number of TSPs"},
	11:  {id: 11, arg1: u8, arg2: u8, desc: "TSP setting"},
	88:  {id: 88, arg1: u8, arg2: nu, desc: "Number of TSPs V/H"},
	89:  {id: 89, arg1: u8, arg2: u8, desc: "TSP setting V/H"},
	105: {id: 105, arg1: u8, arg2: nu, desc: "Number of TSPs solar storage"},
	106: {id: 106, arg1: u8, arg2: u8, desc: "TSP setting solar storage"},
	12:  {id: 12, arg1: u8, arg2: nu, desc: "Size of fault buffer"},
	13:  {id: 13, arg1: u8, arg2: u8, desc: "Fault buffer entry"},
	90:  {id: 90, arg1: u8, arg2: nu, desc: "Size of fault buffer V/H"},
	91:  {id: 91, arg1: u8, arg2: u8, desc: "Fault buffer entry V/H"},
	107: {id: 107, arg1: u8, arg2: u8, desc: "Size of fault buffer solar storage"},
	108: {id: 108, arg1: u8, arg2: u8, desc: "Fault buffer entry solar storage"},
	7:   {id: 7, arg1: f8, arg2: ns, desc: "Cooling control signal"},
	9:   {id: 9, arg1: f8, arg2: ns, desc: "Remote override room setpoint"},
	14:  {id: 14, arg1: f8, arg2: ns, desc: "Maximum relative modulation level"},
	15:  {id: 15, arg1: u8, arg2: u8, desc: "Boiler capacity and modulation limits"},
	100: {id: 100, arg1: nu, arg2: flag8, desc: "Remote override function"},
}

func (s Src) String() string {
	switch s {
	case B:
		return "Boiler"
	case R:
		return "Request"
	case T:
		return "Thermostat"
	case A:
		return "Answer"
	}
	return string(s)
}

func (m MessageType) String() string {
	switch m {
	case 0:
		return "Read-Data"
	case 1:
		return "Write-Data"
	case 2:
		return "Inv-Data"
	case 3:
		return "Reserved"
	case 4:
		return "Read-Ack"
	case 5:
		return "Write-Ack"
	case 6:
		return "Data-Inv"
	case 7:
		return "Unk-DataId"
	}
	return "Unknown"
}

func (m *Message) IsMasterCH() bool          { return m.Values[0].(int64)&int64(MasterCH) > 0 }
func (m *Message) IsMasterDHW() bool         { return m.Values[0].(int64)&int64(MasterDHW) > 0 }
func (m *Message) IsMasterCooling() bool     { return m.Values[0].(int64)&int64(MasterCooling) > 0 }
func (m *Message) IsMasterOTC() bool         { return m.Values[0].(int64)&int64(MasterOTC) > 0 }
func (m *Message) IsMasterCH2() bool         { return m.Values[0].(int64)&int64(MasterCH2) > 0 }
func (m *Message) IsMasterSummer() bool      { return m.Values[0].(int64)&int64(MasterSummer) > 0 }
func (m *Message) IsMasterDHWBlocking() bool { return m.Values[0].(int64)&int64(MasterDHWBlocking) > 0 }

func (m *Message) IsSlaveFault() bool       { return m.Values[1].(int64)&int64(SlaveFault) > 0 }
func (m *Message) IsSlaveCH() bool          { return m.Values[1].(int64)&int64(SlaveCH) > 0 }
func (m *Message) IsSlaveDHW() bool         { return m.Values[1].(int64)&int64(SlaveDHW) > 0 }
func (m *Message) IsSlaveFlame() bool       { return m.Values[1].(int64)&int64(SlaveFlame) > 0 }
func (m *Message) IsSlaveCooling() bool     { return m.Values[1].(int64)&int64(SlaveCooling) > 0 }
func (m *Message) IsSlaveCH2() bool         { return m.Values[1].(int64)&int64(SlaveCH2) > 0 }
func (m *Message) IsSlaveDiagnostic() bool  { return m.Values[1].(int64)&int64(SlaveDiagnostic) > 0 }
func (m *Message) IsSlaveElectricity() bool { return m.Values[1].(int64)&int64(SlaveElectricity) > 0 }

func (m *Message) ThermostatSetPoint(value float64) {
	m.Src = T
	m.Type = WriteData
	m.ID = 1

	m.Values = []interface{}{value}
}

func (m *Message) RoomSetPoint(value float64) {
	m.Src = T
	m.Type = WriteData
	m.ID = 16

	m.Values = []interface{}{value}
}

func (m *Message) Encode() (string, error) {
	md, ok := messageDefs[m.ID]
	if !ok {
		return "", fmt.Errorf("message type not found: %d", m.ID)
	}
	res := fmt.Sprintf("%1c%1X0%2X", m.Src, int(m.Type), m.ID)

	args := []argType{md.arg1, md.arg2}
	for i, arg := range args {
		switch arg {
		case u8, s8, u16, flag8:
			res += fmt.Sprintf("%2X", int(m.Values[i].(float64)))
		case f8:
			f := m.Values[i].(float64)
			v1 := int(f)
			v2 := int(f*100) - int(f)*100
			res += fmt.Sprintf("%2X%2X", v1, v2*255/100)
			break
		}
	}

	return res, nil
}

func (m *Message) Decode(msg string) error {
	var src rune
	var kind int
	var id int
	var data int

	if len(msg) != 9 {
		return fmt.Errorf("wrong message format: %s", msg)
	}

	fmt.Sscanf(msg, "%1c%1X0%2X%4X", &src, &kind, &id, &data)

	switch Src(src) {
	case B, T, A, R:
	default:
		return fmt.Errorf("src not supported: %c", src)
	}

	kind &= 7

	md, ok := messageDefs[id]
	if !ok {
		return fmt.Errorf("message type not found: %d", id)
	}

	m.ID = id
	m.Src = Src(src)
	m.Type = MessageType(kind)
	m.Desc = md.desc

	var values []int64
	if md.arg2 == ns {
		values = []int64{int64(data)}
	} else {
		values = []int64{int64(data) >> 8, int64(data) & 0xff}
	}

	arg := md.arg1

	m.Values = m.Values[:0]
	for _, value := range values {
		switch arg {
		case u8, s8, u16, flag8:
			m.Values = append(m.Values, int64(value))
		case f8:
			f := float64(int8(value >> 8))
			f += float64(float64(value&0xff) / 255)
			f = float64(int64(f*100+0.5)) / 100

			m.Values = append(m.Values, f)
		}
		arg = md.arg2
	}

	return nil
}

type oItemKey struct {
	src  Src
	id   int
	kind MessageType
}

type OpenthermItem struct {
	Item item.Item

	flag Flag
}

type openThermMessageHandler struct {
	opentherm *OpenTherm
}

type currentMessageHandler struct {
	opentherm *OpenTherm
}

type returnTempMessageHandler struct {
	opentherm *OpenTherm
}

type pauseStateMessageHandler struct {
	opentherm *OpenTherm
}

type pauseModeListener struct {
	opentherm *OpenTherm
}

type OpenTherm struct {
	sync.RWMutex

	CurrentItem    *item.AnItem
	ReturnTempItem *item.AnItem
	PauseStateItem *item.AnItem
	PauseModeItem  *button.SwitchItem

	id     string
	oItems map[oItemKey]*OpenthermItem

	forceSetPoint float64
	conn          *hmqtt.MQTTConn
}

func (s *OpenTherm) OnValueChange(it item.Item, old string, new string) {
	payload := "on"
	if new != item.ON {
		payload = "off"
	}
	s.conn.Publish(it.GetID(), "otg/in/pause", payload)
}

func (o *OpenTherm) OnMessage(client mqtt.Client, msg mqtt.Message) {
	new := string(msg.Payload())
	new = strings.TrimRight(new, "\r\n")
	if len(new) == 0 {
		return
	}

	omsg := &Message{}
	if err := omsg.Decode(new); err != nil {
		server.Log.Errorf("unable to decode opentherm message %s: %s", new, err)
		return
	}
	server.Log.Debugf("Opentherm message %d from '%s' '%s': %+v(%s) [%s]", omsg.ID, omsg.Src, omsg.Desc, omsg.Values, omsg.Type, new)

	key := oItemKey{
		src:  omsg.Src,
		id:   omsg.ID,
		kind: omsg.Type,
	}

	o.RLock()
	oitem := o.oItems[key]
	o.RUnlock()

	if oitem != nil {
		if omsg.ID == 0 {
			var state bool

			if omsg.Src == B || omsg.Src == A {
				switch oitem.flag {
				case SlaveFault:
					state = omsg.IsSlaveFault()
				case SlaveCH:
					state = omsg.IsSlaveCH()
				case SlaveDHW:
					state = omsg.IsSlaveDHW()
				case SlaveFlame:
					state = omsg.IsSlaveFlame()
				case SlaveCooling:
					state = omsg.IsSlaveCooling()
				case SlaveCH2:
					state = omsg.IsSlaveCH2()
				case SlaveDiagnostic:
					state = omsg.IsSlaveDiagnostic()
				case SlaveElectricity:
					state = omsg.IsSlaveElectricity()
				}
			} else if omsg.Src == T || omsg.Src == R {
				switch oitem.flag {
				case MasterCH:
					state = omsg.IsMasterCH()
				case MasterDHW:
					state = omsg.IsMasterDHW()
				case MasterCooling:
					state = omsg.IsMasterCooling()
				case MasterOTC:
					state = omsg.IsMasterOTC()
				case MasterCH2:
					state = omsg.IsMasterCH2()
				case MasterSummer:
					state = omsg.IsMasterSummer()
				case MasterDHWBlocking:
					state = omsg.IsMasterDHWBlocking()
				}
			}

			if state {
				oitem.Item.SetValue(item.ON)
			} else {
				oitem.Item.SetValue(item.OFF)
			}
		} else {
			value := omsg.Values[0]
			switch value.(type) {
			case int64:
				oitem.Item.SetValue(fmt.Sprintf("%d", value))
			case float64:
				oitem.Item.SetValue(fmt.Sprintf("%.2f", value))
			}
		}
	}
}

func (p *pauseModeListener) OnValueChange(it item.Item, old string, new string) {
	payload := "off"
	if new != item.ON {
		payload = "on"
	}
	p.opentherm.conn.Publish(it.GetID(), "otg/in/pause", payload)
}

func (o *returnTempMessageHandler) OnMessage(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())

	o.opentherm.ReturnTempItem.SetValue(value)
}

func (o *currentMessageHandler) OnMessage(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())

	amp, _ := strconv.ParseFloat(value, 64)
	watt := amp * 220
	if watt < 200 {
		watt = 0
	}

	o.opentherm.CurrentItem.SetValue(fmt.Sprintf("%.2f", watt))
}

func (o *pauseStateMessageHandler) OnMessage(client mqtt.Client, msg mqtt.Message) {
	value := string(msg.Payload())

	if value == "off" {
		value = item.ON
	} else {
		value = item.OFF
	}

	o.opentherm.PauseStateItem.SetValue(value)
}

func (o *OpenTherm) RegisterFlagItem(id, label, unit string, src Src, kind MessageType, flag Flag) item.Item {
	key := oItemKey{
		src:  src,
		id:   0,
		kind: kind,
	}

	o.Lock()
	defer o.Unlock()

	item := &item.AnItem{
		ID:    id,
		Label: label,
		Type:  "state",
		Img:   "switch",
	}
	o.oItems[key] = &OpenthermItem{
		Item: item,
		flag: flag,
	}

	server.Registry.Add(item, o.id)

	return item
}

func (o *OpenTherm) RegisterValueItem(id, label, unit string, src Src, msgID int, kind MessageType) item.Item {
	key := oItemKey{
		src:  src,
		id:   msgID,
		kind: kind,
	}

	o.Lock()
	defer o.Unlock()

	value := value.NewValueItem(id, label, unit)
	o.oItems[key] = &OpenthermItem{
		Item: value,
	}

	server.Registry.Add(value, o.id)

	return value
}

func (o *OpenTherm) RegisterSetPointItem(label string) item.Item {
	value := value.NewValueItem("OPENTHERM_FORCE_SETPOINT", label, "°")
	value.SetValue("17.5")
	value.Type = "range"

	server.Registry.Add(value, o.id)
	value.AddListener(o)

	return value
}

func NewOpenTherm(id string, conn *hmqtt.MQTTConn, topic, currentTopic, returnTopic, pauseTopic string) *OpenTherm {
	o := &OpenTherm{
		id:     id,
		oItems: make(map[oItemKey]*OpenthermItem),
		CurrentItem: &item.AnItem{
			ID:    "OTG_CURRENT",
			Label: "Current",
			Type:  "value",
			Img:   "electricity",
			Unit:  "W",
		},
		ReturnTempItem: &item.AnItem{
			ID:    "OTG_RETURN_TEMPERATURE",
			Label: "Return temperature",
			Type:  "value",
			Img:   "temperature",
			Unit:  "°",
		},
		PauseStateItem: &item.AnItem{
			ID:    "OTG_RELAY_STATE",
			Label: "State",
			Type:  "state",
			Img:   "plug",
		},
		PauseModeItem: &button.SwitchItem{
			AnItem: item.AnItem{
				ID:    "OTG_RELAY_MODE",
				Label: "Mode",
				Type:  "switch",
				Img:   "plug",
			},
		},
		conn: conn,
	}

	o.PauseModeItem.SetValue(item.ON)

	server.Registry.Add(o.CurrentItem, id)
	server.Registry.Add(o.ReturnTempItem, id)
	server.Registry.Add(o.PauseStateItem, id)
	server.Registry.Add(o.PauseModeItem, id)

	conn.Subscribe(topic, o)
	conn.Subscribe(currentTopic, &currentMessageHandler{opentherm: o})
	conn.Subscribe(returnTopic, &returnTempMessageHandler{opentherm: o})
	conn.Subscribe(pauseTopic, &pauseStateMessageHandler{opentherm: o})

	o.PauseModeItem.AddListener(&pauseModeListener{opentherm: o})

	return o
}
