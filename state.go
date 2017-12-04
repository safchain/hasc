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
	"sync"

	"golang.org/x/sync/syncmap"
)

type stateListener struct {
	sync.RWMutex
	callback func(object Object, old string, new string)
	eventMap syncmap.Map
}

func (s *stateListener) OnStateChange(object Object, old string, new string) {
	if _, ok := s.eventMap.Load(object.ID()); ok {
		return
	}

	s.eventMap.Store(object.ID(), true)
	defer func() {
		s.eventMap.Delete(object.ID())
	}()

	s.RLock()
	if s.callback != nil {
		s.callback(object, old, new)
	}
	s.RUnlock()
}

func (s *stateListener) setCallbackFnc(cb func(object Object, old string, new string)) {
	s.Lock()
	s.callback = cb
	s.Unlock()
}
