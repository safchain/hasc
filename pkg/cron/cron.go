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

package cron

import (
	"log"

	"github.com/robfig/cron"

	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/server"
)

type Cron struct {
}

type CronOpts struct {
	Value string
}

// NewCronTrigger creates a new Unix like cron job with the given cron expression (man 5 crontab)
// It will change the state of the given object. By default the object state will be set
// to ON. The state used can be changed using the CronOpts parameter.
func NewCronTrigger(schedule string, it item.Item, opts ...CronOpts) *Cron {
	c := cron.New()

	value := item.ON
	if len(opts) > 0 && opts[0].Value != "" {
		value = opts[0].Value
	}

	err := c.AddFunc(schedule, func() {
		it.SetValue(value)
	})
	if err != nil {
		log.Fatalf("Cron unable to parse schedule definition: %s", err)
	}

	server.Log.Infof("New Cron %s for %s", schedule, it.GetID())

	c.Start()

	return &Cron{}
}
