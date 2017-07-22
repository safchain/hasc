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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

type eventGCal struct {
	id          string
	summary     string
	description string
	cancel      chan bool
}

type GCal struct {
	sync.RWMutex
	service *calendar.Service
	events  map[string]*eventGCal
}

func (e *eventGCal) stop() {
	e.cancel <- true
}

func eventGCalID(event *calendar.Event) string {
	u := uuid.NewV5(uuid.NamespaceOID,
		fmt.Sprintf("%s-%s-%s-%s-%s-%s", event.Summary, event.Description,
			event.Start.Date, event.Start.DateTime,
			event.End.Date, event.End.DateTime))
	return u.String()
}

func (g *GCal) newEventGCal(event *calendar.Event) (*eventGCal, error) {
	e := &eventGCal{
		id:          eventGCalID(event),
		summary:     event.Summary,
		description: event.Description,
		cancel:      make(chan bool, 2),
	}

	var start, end time.Time
	var err error

	if event.Start.DateTime != "" {
		start, err = time.Parse(time.RFC3339, event.Start.DateTime)
	} else {
		start, err = time.Parse(time.RFC3339, event.Start.Date)
	}
	if err != nil {
		return nil, fmt.Errorf("GCal unable to parse event date: %v", event)
	}

	if event.End.DateTime != "" {
		end, err = time.Parse(time.RFC3339, event.End.DateTime)
	} else {
		end, err = time.Parse(time.RFC3339, event.End.Date)
	}
	if err != nil {
		return nil, fmt.Errorf("GCal unable to parse event date: %v", event)
	}

	now := time.Now()
	if start.Before(now) {
		return nil, fmt.Errorf("GCal event start in the past: %v", event)
	}

	startIn := start.Sub(now)
	endIn := end.Sub(now)

	startAfter := time.After(startIn)
	endAfter := time.After(endIn)

	go func() {
		defer func() {
			g.Lock()
			delete(g.events, e.id)
			g.Unlock()
		}()

		for {
			select {
			case <-startAfter:
				re := regexp.MustCompile("(?i)START:\\s*([^\\s]*)\\s*(.*)")
				if res := re.FindStringSubmatch(event.Description); len(res) > 0 {
					if obj := ObjectFromID(res[1]); obj != nil {

						obj.SetState(res[2])
						Log.Infof("GCal set %s to %s", obj.ID(), res[2])
					}
				}
			case <-endAfter:
				re := regexp.MustCompile("(?i)END:\\s*([^\\s]*)\\s*(.*)")
				if res := re.FindStringSubmatch(event.Description); len(res) > 0 {
					if obj := ObjectFromID(res[1]); obj != nil {
						Log.Infof("GCal set %s to %s", obj.ID(), res[2])
						obj.SetState(res[2])
					}
				}
				return
			case <-e.cancel:
				Log.Infof("GCal event terminated: %s summary: %s, description: %s", e.id, e.summary, strings.Replace(e.description, "\n", "; ", -1))
				return
			}
		}
	}()

	return e, nil
}

func (g *GCal) scheduleGCalEvent(event *calendar.Event) (*eventGCal, error) {
	g.RLock()
	e, ok := g.events[event.Id]
	g.RUnlock()

	if ok {
		if e.id == eventGCalID(event) {
			return e, nil
		}
		e.stop()
	}
	e, err := g.newEventGCal(event)
	if err != nil {
		return nil, err
	}

	g.Lock()
	g.events[event.Id] = e
	g.Unlock()

	Log.Infof("GCal new event scheduled: %s summary: %s, description: %s", event.Id, event.Summary, strings.Replace(event.Description, "\n", "; ", -1))

	return e, nil
}

func (g *GCal) tokenCacheFile() string {
	return filepath.Join(Cfg.GetString("data"), url.QueryEscape("gcal_token.json"))
}

func (g *GCal) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	Log.Warningf("GCal go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		Log.Errorf("GCal unable to read authorization code %v", err)
		return nil
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("GCal unable to retrieve token from web %v", err)
		return nil
	}
	return tok
}

func (g *GCal) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func (g *GCal) saveToken(file string, token *oauth2.Token) {
	Log.Infof("GCal saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		Log.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (g *GCal) refreshFnc(name string) {
	Log.Infof("GCal %s refresh", name)
	l, err := g.service.CalendarList.List().Do()
	if err != nil {
		Log.Errorf("GCal unable to list calendars: %s", err)
		return
	}

	var item *calendar.CalendarListEntry
	for _, i := range l.Items {
		if i.Summary == name {
			item = i
			break
		}
	}

	if item == nil {
		Log.Errorf("GCal calendar %s not found", name)
		return
	}

	t := time.Now().Format(time.RFC3339)
	events, err := g.service.Events.List(item.Id).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		Log.Errorf("GCal unable to retrieve next ten of the user's events. %v", err)
		return
	}

	scheduled := make(map[string]*eventGCal)
	for _, i := range events.Items {
		summary := strings.ToLower(i.Summary)
		if strings.Index(summary, "hasc:") < 0 {
			continue
		}

		e, err := g.scheduleGCalEvent(i)
		if err != nil {
			Log.Errorf("GCal error while scheduling: %s", err)
			continue
		}
		scheduled[e.id] = e
	}

	// close event not scheduled anymore
	g.RLock()
	for _, s := range g.events {
		if _, ok := scheduled[s.id]; !ok {
			s.stop()
		}
	}
	g.RUnlock()
}

func (g *GCal) refresh(name string, refresh time.Duration) {
	g.refreshFnc(name)

	ticker := time.NewTicker(refresh)
	for range ticker.C {
		g.refreshFnc(name)
	}
}

func (g *GCal) getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	tok, err := g.tokenFromFile(g.tokenCacheFile())
	if err != nil {
		tok = g.getTokenFromWeb(config)
		g.saveToken(g.tokenCacheFile(), tok)
	}
	return config.Client(ctx, tok)
}

// NewGCalTrigger creates a GCal trigger allowing to trigger state changes according
// to rules present in GCal events. It will use the gcal_secret.json in the data
// folder. Please go to https://console.developers.google.com for further explanation.
// The format that has to be used is the following :
// START: <Object ID> <State>
// END: <Object ID> <State>
func NewGCalTrigger(name string, refresh time.Duration) *GCal {
	Log.Infof("New GCal: %s", name)

	secretFile := filepath.Join(Cfg.GetString("data"), url.QueryEscape("gcal_secret.json"))
	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	g := &GCal{
		events: make(map[string]*eventGCal),
	}

	client := g.getClient(context.Background(), config)
	if g.service, err = calendar.New(client); err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	go g.refresh(name, refresh)

	return g
}
