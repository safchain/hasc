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
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/safchain/hasc/statics"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type eventListener struct {
}

type row struct {
	Label string
	Img   string
	Items []Item
}

type layout struct {
	rows []row
}

type wsclient struct {
	sync.RWMutex
	addr     net.Addr
	lastRead time.Time
	wch      chan string
	quit     chan bool
}

var (
	// Layout layout provided by the server.
	Layout layout
	// KV Key/Value data store.
	KV *kvStore
	// Log logging instance.
	Log = logging.MustGetLogger("default")
	// Cmd command line parser.
	Cmd *cobra.Command
	// Cfg config file parser.
	Cfg *viper.Viper

	listener  eventListener
	state     stateListener
	objects   map[string]Object
	router    *mux.Router
	wsclients map[*wsclient]*wsclient
	lock      sync.RWMutex
)

func (r *row) MarshalJSON() ([]byte, error) {
	if r.Label != "" {
		return json.Marshal(&struct {
			Label string `json:"label"`
			Img   string `json:"img"`
			Items []Item `json:"items"`
		}{
			Label: r.Label,
			Img:   r.Img,
			Items: r.Items,
		})
	}
	return json.Marshal(r.Items[0])
}

// RegisterObject registers the given object.
func RegisterObject(o Object) Object {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := objects[o.ID()]; ok {
		Log.Fatalf("can't register two object with the same ID: %s", o.ID())
	}

	Log.Infof("Register new %T: %s(%s)", o, o.ID(), o.Label())
	objects[o.ID()] = o

	// listen for item object changes
	o.AddObjectListener(listener)
	o.AddObjectListener(&state)

	return o
}

// ObjectFromID looks up for an registered object for the given id.
func ObjectFromID(id string) Object {
	lock.RLock()
	defer lock.RUnlock()

	o, ok := objects[id]
	if !ok {
		Log.Errorf("Object ID not found: %s", id)

		// return fake object so that no need to check nil value
		return &AnObject{}
	}
	return o
}

// SetObjectIDState set state of the given object id
func SetObjectIDState(id string, state string) {
	ObjectFromID(id).SetState(state)
}

func Objects() []Object {
	lock.RLock()
	defer lock.RUnlock()

	var l []Object
	for _, o := range objects {
		l = append(l, o)
	}
	return l
}

// SetStateListener set a state listener. It is useful to implement gateways for
// example leveraging a MQTT broker.
func SetStateListener(fnc func(object Object, old string, new string)) {
	state.setCallbackFnc(fnc)
}

func getObjectState(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	object := ObjectFromID(params["id"])
	w.Write([]byte(object.State()))
}

func setObjectState(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	data, _ := ioutil.ReadAll(r.Body)
	ObjectFromID(params["id"]).SetState(string(data))
}

func asset(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if strings.HasPrefix(upath, "/") {
		upath = strings.TrimPrefix(upath, "/")
	}

	content, err := statics.Asset(upath)
	if err != nil {
		Log.Errorf("unable to find the asset: %s", upath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext := filepath.Ext(upath)
	ct := mime.TypeByExtension(ext)

	w.Header().Set("Content-Type", ct+"; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func indexHTML(w http.ResponseWriter, r *http.Request) {
	asset := statics.MustAsset("statics/server.html")

	header := true
	if r.FormValue("header") == "false" {
		header = false
	}

	data := struct {
		Rows   []row
		Header bool
	}{
		Rows:   Layout.rows,
		Header: header,
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.New("index").Parse(string(asset)))
	if err := tmpl.Execute(w, data); err != nil {
		Log.Criticalf("template execution error: %s", err)
	}
}

func indexJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")

	o := struct {
		HascVer string
		Objects []row
	}{
		HascVer: Version,
		Objects: Layout.rows,
	}

	e.Encode(o)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("type") == "json" {
		indexJSON(w, r)
	} else {
		indexHTML(w, r)
	}
}

func websocket(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		Log.Errorf("Websocket error: %s", err)
		return
	}
	Log.Infof("websocket new client from: %s", r.Host)

	// resend all item values
	for _, row := range Layout.rows {
		for _, item := range row.Items {
			msg := fmt.Sprintf("%s_%s=%s", item.Object().ID(), item.ID(), item.Value())
			Log.Infof("websocket send message: %s", msg)

			err = wsutil.WriteServerMessage(conn, ws.OpText, []byte(msg))
			if err != nil {
				Log.Warningf("websocket error while writing message: %s", err)
			}

			if _, ok := item.(*GroupItem); ok {
				for _, subitem := range item.Object().Items() {
					msg := fmt.Sprintf("%s_%s=%s", subitem.Object().ID(), subitem.ID(), subitem.Value())
					Log.Infof("websocket send message: %s", msg)

					err = wsutil.WriteServerMessage(conn, ws.OpText, []byte(msg))
					if err != nil {
						Log.Warningf("websocket error while writing message: %s", err)
					}
				}
			}
		}
	}

	client := &wsclient{
		addr:     conn.RemoteAddr(),
		lastRead: time.Now(),
		wch:      make(chan string, 100),
	}

	lock.Lock()
	wsclients[client] = client
	lock.Unlock()

	go func() {
		for {
			_, _, err = wsutil.ReadClientData(conn)
			if err != nil {
				return
			}

			client.Lock()
			client.lastRead = time.Now()
			client.Unlock()
		}
	}()

	go func() {
		defer func() {
			conn.Close()

			lock.Lock()
			delete(wsclients, client)
			close(client.wch)
			lock.Unlock()

			Log.Infof("websocket client removed: %s", client.addr)
		}()

		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		for {
			select {
			case msg := <-client.wch:
				err = wsutil.WriteServerMessage(conn, ws.OpText, []byte(msg))
				if err != nil {
					Log.Warningf("websocket error while writing message: %s", err)
					return
				}
			case now := <-tick.C:
				client.RLock()
				out := client.lastRead.Add(10 * time.Second).Before(now)
				client.RUnlock()

				if out {
					return
				}
			}
		}
	}()
}

func notifyItemStateChange(item Item) {
	for _, client := range wsclients {
		msg := fmt.Sprintf("%s_%s=%s", item.Object().ID(), item.ID(), item.Value())
		Log.Infof("websocket send message to %s: %s", client.addr, msg)
		client.wch <- msg
	}
}

func (l eventListener) OnStateChange(object Object, old string, new string) {
	lock.RLock()
	defer lock.RUnlock()

	for _, row := range Layout.rows {
		for _, item := range row.Items {
			if item.Object().ID() == object.ID() {
				notifyItemStateChange(item)
			}
			if _, ok := item.(*GroupItem); ok {
				for _, subitem := range item.Object().Items() {
					if subitem.Object().ID() == object.ID() {
						notifyItemStateChange(subitem)
					}
				}
			}
		}
	}
}

// AddItem adds the given item to the server layout.
func (l *layout) AddItem(item Item) {
	lock.Lock()
	defer lock.Unlock()

	l.rows = append(l.rows, row{Items: []Item{item}})
}

// AddItems adds the items to the server layout. The given label, image will be
// used to render the item group.
func (l *layout) AddItems(label string, img string, items ...Item) {
	lock.Lock()
	defer lock.Unlock()

	l.rows = append(l.rows, row{Label: label, Img: img, Items: items})
}

func listenAndServe() {
	Log.Info("Hasc server started")
	Log.Fatal(http.ListenAndServe(":"+Cfg.GetString("port"), nil))
}

func init() {
	Cmd = &cobra.Command{}
}

// Start start the server. It starts the WebSocket server, listens for objects state
// changes and forward them trough WebSocket. The onInit callback has to be used
// to register objects.
func Start(name string, onInit func()) {
	format := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} ▶ %{level:.6s}%{color:reset} %{message}`)
	logging.SetFormatter(format)

	defaultPort := "12345"
	var defaultDataDir string

	usr, err := user.Current()
	if err == nil {
		defaultDataDir = filepath.Join(usr.HomeDir, ".hasc")
	} else {
		Log.Warning("unable to get current user path, defaulting to /var/lib/hasc")
		defaultDataDir = "/var/lib/hasc"
	}

	var cfgFile, port, data string

	Cmd.Use = name
	Cmd.Short = fmt.Sprintf("%s based on H.A.S.C", name)
	Cmd.Long = fmt.Sprintf(`%s based on H.A.S.C.
H.A.S.C. is a home automation framework written in GO.
Complete documentation is available at http://github.com/safchain/hasc`, name)
	Cmd.Run = func(cmd *cobra.Command, args []string) {
		if cfgFile != "" {
			Cfg.SetConfigFile(cfgFile)
			if err := Cfg.ReadInConfig(); err != nil {
				fmt.Println("can't read config: ", err)
				os.Exit(1)
			}
		}

		data = Cfg.GetString("data")
		if err := os.MkdirAll(data, 0700); err != nil {
			fmt.Println("unable to create data dir: ", err)
			os.Exit(1)
		}

		router = mux.NewRouter()
		router.HandleFunc("/", index).Methods("GET")
		router.PathPrefix("/statics").HandlerFunc(asset).Methods("GET")
		router.HandleFunc("/object/{id}", getObjectState).Methods("GET")
		router.HandleFunc("/object/{id}", setObjectState).Methods("POST")
		router.HandleFunc("/ws", websocket)

		if Cfg.GetString("password") != "" {
			http.Handle("/", httpauth.SimpleBasicAuth(Cfg.GetString("username"), Cfg.GetString("password"))(router))
		} else {
			http.Handle("/", router)
		}

		objects = make(map[string]Object)
		wsclients = make(map[*wsclient]*wsclient)

		KV = newKVStore()

		onInit()

		listenAndServe()
	}

	Cmd.PersistentFlags().StringVarP(&cfgFile, "conf", "", "", "config file (optional)")
	Cmd.PersistentFlags().StringVarP(&port, "port", "", defaultPort, "port (default is 12345)")
	Cmd.PersistentFlags().StringVarP(&data, "data", "", defaultDataDir, "data dir (default is $HOME/.hasc)")

	Cfg = viper.New()
	Cfg.BindPFlag("port", Cmd.PersistentFlags().Lookup("port"))
	Cfg.BindPFlag("data", Cmd.PersistentFlags().Lookup("data"))
	Cfg.SetDefault("port", defaultPort)
	Cfg.SetDefault("data", defaultDataDir)
	Cfg.SetDefault("username", "admin")

	Cmd.Execute()
}
