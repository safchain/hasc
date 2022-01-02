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

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/safchain/hasc/pkg/influxdb"
	"github.com/safchain/hasc/pkg/item"
	"github.com/safchain/hasc/pkg/kv"
	"github.com/safchain/hasc/pkg/registry"
	"github.com/safchain/hasc/pkg/version"
	"github.com/safchain/hasc/statics"
)

type itemListener struct{}

type row struct {
	Item     item.Item
	SubItems []item.Item
}

type layout struct {
	rows []row
}

type wsclient struct {
	sync.RWMutex
	addr     net.Addr
	lastRead time.Time
	wch      chan []byte
}

var (
	// Layout layout provided by the server.
	Layout layout
	// KV Key/Value data store.
	KV *kv.KVStore
	// Log logging instance.
	Log = logging.MustGetLogger("default")
	// Cmd command line parser.
	Cmd *cobra.Command
	// Cfg config file parser.
	Cfg *viper.Viper
	// Registry item registry
	Registry *registry.Registry

	// InfluxDB database
	InfluxDB *influxdb.InfluxDB

	listener  itemListener
	router    *mux.Router
	wsclients map[*wsclient]*wsclient
	lock      sync.RWMutex
)

func getItemValue(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]
	if params["subid"] != "" {
		id += "/" + params["subid"]
	}

	item := Registry.Get(id)
	if item != nil {
		w.Write([]byte(item.GetValue()))
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func getItemValues(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]
	if params["subid"] != "" {
		id += "/" + params["subid"]
	}

	item := Registry.Get(id)
	if item == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)

		return
	}
	values := InfluxDB.GetValues(item)
	b, err := json.Marshal(values)
	if err != nil {
		Log.Errorf("error while marshalling values of %s: %s", item.GetID(), err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))

		return
	}

	w.Write(b)

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func setItemValue(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]
	if params["subid"] != "" {
		id += "/" + params["subid"]
	}

	item := Registry.Get(id)
	if item != nil {
		data, _ := ioutil.ReadAll(r.Body)
		item.SetValue(string(data))
	}
}

func asset(path string, w http.ResponseWriter, r *http.Request) {
	path = "statics/" + path

	content, err := statics.Asset(path)
	if err != nil {
		Log.Errorf("unable to find the asset: %s", path)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext := filepath.Ext(path)
	ct := mime.TypeByExtension(ext)

	w.Header().Set("Content-Type", ct+"; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if strings.HasPrefix(upath, "/") {
		upath = strings.TrimPrefix(upath, "/")
	}
	asset(upath, w, r)
}

func indexJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")

	o := struct {
		HascVer string
		Rows    []row
	}{
		HascVer: version.Version,
		Rows:    Layout.rows,
	}

	e.Encode(o)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("type") == "json" {
		indexJSON(w, r)
	} else {
		asset("index.html", w, r)
	}
}

func websocket(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		Log.Errorf("Websocket error: %s", err)
		return
	}
	Log.Infof("websocket new client from: %s", r.Host)

	for _, item := range Registry.Items() {
		b, err := json.Marshal(item)
		if err != nil {
			Log.Errorf("websocket error while writing message: %s", err)
			continue
		}
		Log.Infof("websocket send message: %s", string(b))

		err = wsutil.WriteServerMessage(conn, ws.OpText, b)
		if err != nil {
			Log.Warningf("websocket error while writing message: %s", err)
			continue
		}
	}

	client := &wsclient{
		addr:     conn.RemoteAddr(),
		lastRead: time.Now(),
		wch:      make(chan []byte, 1000),
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
			case b := <-client.wch:
				if err = wsutil.WriteServerMessage(conn, ws.OpText, b); err != nil {
					Log.Warningf("websocket error while writing message: %s", err)
					return
				}
			case now := <-tick.C:
				client.RLock()
				out := client.lastRead.Add(30 * time.Second).Before(now)
				client.RUnlock()

				if out {
					return
				}
			}
		}
	}()
}

func (l itemListener) OnValueChange(it item.Item, old string, new string) {
	lock.RLock()
	for _, client := range wsclients {
		b, err := json.Marshal(it)
		if err != nil {
			Log.Errorf("websocket error while writing message: %s", err)
		}

		Log.Infof("websocket send message to %s: %s", client.addr, string(b))
		client.wch <- b
	}
	lock.RUnlock()
}

func (l *layout) AddItem(it item.Item) {
	lock.Lock()
	defer lock.Unlock()

	l.rows = append(l.rows, row{Item: it})
}

func (l *layout) AddItems(it item.Item, subItems ...item.Item) {
	lock.Lock()
	defer lock.Unlock()

	l.rows = append(l.rows, row{Item: it, SubItems: subItems})
}

func GetItem(id string) item.Item {
	it := Registry.Get(id)
	if it == nil {
		return &item.AnItem{}
	}
	return it
}

func listenAndServe(router *mux.Router) {
	server := &http.Server{Addr: ":" + Cfg.GetString("port"), Handler: router}
	server.SetKeepAlivesEnabled(false)
	server.ListenAndServe()

	Log.Infof("Hasc server started, listen: %s", Cfg.GetString("port"))
	Log.Fatal(server.ListenAndServe())
}

func init() {
	Cmd = &cobra.Command{}
}

type Auth struct {
	stdHandler  http.Handler
	authHandler http.Handler
}

func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		a.stdHandler.ServeHTTP(w, r)
	} else {
		a.authHandler.ServeHTTP(w, r)
	}
}

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
		router.PathPrefix("/static").HandlerFunc(assetHandler).Methods("GET")
		router.PathPrefix("/statics").HandlerFunc(assetHandler).Methods("GET")
		router.HandleFunc("/item/{id}", getItemValue).Methods("GET")
		router.HandleFunc("/item/{id}", setItemValue).Methods("POST")
		router.HandleFunc("/item/{id}/{subid}", getItemValue).Methods("GET")
		router.HandleFunc("/item/{id}/{subid}", setItemValue).Methods("POST")
		router.HandleFunc("/values/{id}", getItemValues).Methods("GET")
		router.HandleFunc("/values/{id}/{subid}", getItemValues).Methods("GET")
		router.HandleFunc("/ws", websocket)

		headersOk := handlers.AllowedHeaders([]string{"Authorization"})
		originsOk := handlers.AllowedOrigins([]string{"*"})
		methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

		handler := handlers.CORS(originsOk, headersOk, methodsOk)(router)

		if Cfg.GetString("password") != "" {
			handler = &Auth{
				stdHandler:  handler,
				authHandler: httpauth.SimpleBasicAuth(Cfg.GetString("username"), Cfg.GetString("password"))(handler),
			}
		}
		http.Handle("/", handler)

		Registry = registry.NewRegistry(listener)

		wsclients = make(map[*wsclient]*wsclient)

		KV = kv.NewKVStore(Cfg)

		InfluxDB = influxdb.NewInfluxDB(Cfg, Log)

		onInit()

		listenAndServe(router)
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
