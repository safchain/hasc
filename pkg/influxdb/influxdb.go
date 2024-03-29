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

package influxdb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/op/go-logging"
	"github.com/spf13/viper"

	"github.com/safchain/hasc/pkg/item"
)

type InfluxDB struct {
	sync.RWMutex
	client    influxdb.Client
	db        string
	bp        influxdb.BatchPoints
	cq        map[string]bool
	flush     time.Duration
	lastFlush time.Time
	points    chan *Point
	logger    *logging.Logger
}

type Point struct {
	id    string
	tags  map[string]string
	value float64
}

func (i *InfluxDB) createCQs(name string) error {
	cq := `CREATE CONTINUOUS QUERY "%s.1d" ON "%s" BEGIN
		SELECT mean("value") AS "value"
		INTO "1d"."%s"
		FROM "%s"
		GROUP BY time(5m)
	END`

	if err := i.rawQuery(fmt.Sprintf(cq, name, i.db, name, name)); err != nil {
		return err
	}

	cq = `CREATE CONTINUOUS QUERY "%s.1w" ON "%s" BEGIN
		SELECT mean("value") AS "value"
		INTO "1w"."%s"
		FROM "1d"."%s"
		GROUP BY time(30m)
	END`

	if err := i.rawQuery(fmt.Sprintf(cq, name, i.db, name, name)); err != nil {
		return err
	}

	cq = `CREATE CONTINUOUS QUERY "%s.1y" ON "%s" BEGIN
		SELECT mean("value") AS "value"
		INTO "1y"."%s"
		FROM "1w"."%s"
		GROUP BY time(2h)
	END`

	if err := i.rawQuery(fmt.Sprintf(cq, name, i.db, name, name)); err != nil {
		return err
	}

	return nil
}

func (i *InfluxDB) OnValueChange(it item.Item, old string, new string) {
	id := it.GetID()
	i.logger.Infof("InfluxDB insert data points for %s", id)

	tags := make(map[string]string)

	var f float64
	var err error

	value := it.GetValue()
	switch value {
	case "on", "ON":
		f = 1.0
	case "off", "OFF":
		f = 0.0
	default:
		f, err = strconv.ParseFloat(value, 64)
		if err != nil {
			i.logger.Errorf("InfluxDB value %s is not a numeric: %s", it.GetValue(), err)
			return
		}
	}

	select {
	case i.points <- &Point{id: id, tags: tags, value: f}:
	default:
	}
}

func (i *InfluxDB) insertPoints() {
	for point := range i.points {
		fields := map[string]interface{}{
			"value": point.value,
		}

		/*if _, ok := i.cq[id]; !ok {
			if err = i.createCQs(id); err != nil {
				Log.Errorf("InfluxDB unable to create CQs for %s: %s", id, err)
				return
			}
			i.cq[id] = true
		}*/

		pt, err := influxdb.NewPoint(point.id, point.tags, fields, time.Now())
		if err != nil {
			i.logger.Errorf("InfluxDB new point error: %s", err)
			continue
		}

		i.addPoint(pt)
	}
}

func (i *InfluxDB) addPoint(pt *influxdb.Point) {
	i.Lock()
	defer i.Unlock()

	if i.client == nil {
		return
	}

	if i.bp == nil {
		bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
			Database:  i.db,
			Precision: "s",
		})
		if err != nil {
			i.logger.Fatalf("InfluxDB new batch points error: %s", err)
		}
		i.bp = bp
	}
	i.bp.AddPoint(pt)

	if i.lastFlush.Add(i.flush).After(time.Now()) {
		return
	}
	i.logger.Infof("InfluxDB flush data points")

	err := i.client.Write(i.bp)
	if err != nil {
		i.logger.Errorf("InfluxDB write error: %s", err)
	}

	i.lastFlush = time.Now()
	i.bp = nil
}

func (i *InfluxDB) rawQuery(q string) error {
	i.Lock()
	defer i.Unlock()

	if i.client == nil {
		return nil
	}

	query := influxdb.Query{
		Command: q,
	}

	resp, err := i.client.Query(query)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error()
	}

	return nil
}

func (i *InfluxDB) createDatabase() error {
	if err := i.rawQuery(fmt.Sprintf("CREATE DATABASE %s", i.db)); err != nil {
		return err
	}

	/*q := fmt.Sprintf(`CREATE RETENTION POLICY "1d" ON "%s" DURATION 1d REPLICATION 1 DEFAULT`, i.db)
	if err := i.rawQuery(q); err != nil {
		return err
	}

	q = fmt.Sprintf(`CREATE RETENTION POLICY "1w" ON "%s" DURATION 1w REPLICATION 1`, i.db)
	if err := i.rawQuery(q); err != nil {
		return err
	}

	q = fmt.Sprintf(`CREATE RETENTION POLICY "1y" ON "%s" DURATION 52w REPLICATION 1`, i.db)
	if err := i.rawQuery(q); err != nil {
		return err
	}*/

	return nil
}

func (i *InfluxDB) GetValues(item item.Item) [][]interface{} {
	query := influxdb.Query{
		Command:   fmt.Sprintf(`SELECT mean("value") FROM "%s" WHERE time >= now() - 3h GROUP BY time(60s) fill(previous)`, item.GetID()),
		Database:  i.db,
		Precision: "ms",
	}

	fmt.Println(query.Command)

	resp, err := i.client.Query(query)
	if err != nil {
		return nil
	}

	if len(resp.Results) == 0 || len(resp.Results[0].Series) == 0 {
		return nil
	}

	var values [][]interface{}

	for _, value := range resp.Results[0].Series[0].Values {
		ms, _ := value[0].(json.Number).Int64()
		date := time.Unix(ms/1000, 0)

		values = append(values, []interface{}{
			fmt.Sprintf("%02d:%02d", date.Hour(), date.Minute()),
			value[1],
		})
	}

	return values
}

func (i *InfluxDB) Watch(item item.Item) {
	item.AddListener(i)
	item.EnableHistory()
}

// NewInfluxDB returns a new instance of influxdb time series database. It implements
// the ObjectListener interface thus It will store the states objects monitored.
func NewInfluxDB(cfg *viper.Viper, logger *logging.Logger) *InfluxDB {
	addr := cfg.GetString("influxdb.addr")
	port := cfg.GetInt("influxdb.port")
	db := cfg.GetString("influxdb.db")
	username := cfg.GetString("influxdb.username")
	password := cfg.GetString("influxdb.password")
	flush := cfg.GetInt("influxdb.flush")

	i := &InfluxDB{
		db:        db,
		flush:     time.Duration(time.Second * time.Duration(flush)),
		cq:        make(map[string]bool),
		lastFlush: time.Now(),
		points:    make(chan *Point, 100000),
		logger:    logger,
	}

	go func() {
		for {
			time.Sleep(2 * time.Second)

			c, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
				Addr:     fmt.Sprintf("http://%s:%d", addr, port),
				Username: username,
				Password: password,
			})
			if err != nil {
				i.logger.Errorf("InfluxDB new client error: %s", err)
				continue
			}

			i.client = c

			if err := i.createDatabase(); err != nil {
				i.logger.Errorf("InfluxDB unable to create the database: %s", err)
				continue
			}

			break
		}

		i.insertPoints()
	}()

	return i
}
