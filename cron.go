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
	"log"

	"github.com/robfig/cron"
)

type Cron struct {
}

type CronOpts struct {
	State string
}

// NewCronTrigger creates a new Unix like cron job with the given cron expression (man 5 crontab)
// It will change the state of the given object. By default the object state will be set
// to ON. The state used can be changed using the CronOpts parameter.
func NewCronTrigger(schedule string, obj Object, opts ...CronOpts) *Cron {
	c := cron.New()

	state := ON
	if len(opts) > 0 && opts[0].State != "" {
		state = opts[0].State
	}

	err := c.AddFunc(schedule, func() {
		obj.SetState(state)
	})
	if err != nil {
		log.Fatalf("Cron unable to parse schedule definition: %s", err)
	}

	Log.Infof("New Cron %s for %s", schedule, obj.ID())

	c.Start()

	return &Cron{}
}
