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
	"strconv"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

type InfluxDB struct {
	sync.RWMutex
	client    influxdb.Client
	db        string
	bp        influxdb.BatchPoints
	cq        map[string]bool
	flush     time.Duration
	lastFlush time.Time
	points    chan Point
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

func (i *InfluxDB) OnValueChange(item Item, old string, new string) {
	id := item.ID()
	Log.Infof("InfluxDB insert data points for %s", id)

	tags := make(map[string]string)

	var f float64
	var err error

	value := item.Value()
	switch value {
	case "on", "ON":
		f = 1.0
	case "off", "OFF":
		f = 0.0
	default:
		f, err = strconv.ParseFloat(value, 64)
		if err != nil {
			Log.Errorf("InfluxDB value %s is not a numeric: %s", item.Value(), err)
			return
		}
	}

	select {
	case i.points <- Point{id: id, tags: tags, value: f}:
	default:
	}
}

func (i *InfluxDB) insertPoints() {
	for {
		select {
		case point := <-i.points:
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
				Log.Errorf("InfluxDB new point error: %s", err)
				continue
			}

			i.addPoint(pt)
		}
	}
}

func (i *InfluxDB) addPoint(pt *influxdb.Point) {
	i.Lock()
	defer i.Unlock()

	if i.bp == nil {
		bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
			Database:  i.db,
			Precision: "s",
		})
		if err != nil {
			Log.Fatalf("InfluxDB new batch points error: %s", err)
		}
		i.bp = bp
	}
	i.bp.AddPoint(pt)

	if i.lastFlush.Add(i.flush).After(time.Now()) {
		return
	}
	Log.Infof("InfluxDB flush data points")

	err := i.client.Write(i.bp)
	if err != nil {
		Log.Errorf("InfluxDB write error: %s", err)
	}

	i.lastFlush = time.Now()
	i.bp = nil
}

func (i *InfluxDB) rawQuery(q string) error {
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

func (i *InfluxDB) Watch(item Item) {
	item.AddListener(i)
}

// NewInfluxDB returns a new instance of influxdb time series database. It implements
// the ObjectListener interface thus It will store the states objects monitored.
func NewInfluxDB(addr string, port int, db string, username string, password string, flush time.Duration) *InfluxDB {
	c, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%d", addr, port),
		Username: username,
		Password: password,
	})
	if err != nil {
		Log.Fatalf("InfluxDB new client error: %s", err)
	}

	i := &InfluxDB{
		client:    c,
		db:        db,
		flush:     flush,
		cq:        make(map[string]bool),
		lastFlush: time.Now(),
		points:    make(chan Point, 100),
	}

	if err := i.createDatabase(); err != nil {
		Log.Fatalf("InfluxDB unable to create the database: %s", err)
	}

	go i.insertPoints()

	return i
}
