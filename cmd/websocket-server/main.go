// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/illacloud/builder-backend/internal/idconvertor"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/internal/util"
	ws "github.com/illacloud/builder-backend/internal/websocket"

	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/db"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/illacloud/builder-backend/pkg/state"
	filter "github.com/illacloud/builder-backend/pkg/websocket-filter"

	"github.com/gorilla/mux"
	gws "github.com/gorilla/websocket"
)

// websocket client hub

var tssi *state.TreeStateServiceImpl
var kvssi *state.KVStateServiceImpl
var sssi *state.SetStateServiceImpl
var asi *app.AppServiceImpl
var rsi *resource.ResourceServiceImpl
var treeStateRepositoryImpl *repository.TreeStateRepositoryImpl
var kvstateRepositoryImpl *repository.KVStateRepositoryImpl
var setStateRepositoryImpl *repository.SetStateRepositoryImpl
var actionRepositoryImpl *repository.ActionRepositoryImpl
var appRepositoryImpl *repository.AppRepositoryImpl
var appSnapshotRepositoryImpl *repository.AppSnapshotRepositoryImpl

func initEnv() error {
	sugaredLogger := util.NewSugardLogger()
	dbConfig, err := db.GetConfig()
	if err != nil {
		return err
	}
	gormDB, err := db.NewDbConnection(dbConfig, sugaredLogger)
	if err != nil {
		return err
	}

	// init repo
	resourceRepositoryImpl := repository.NewResourceRepositoryImpl(sugaredLogger, gormDB)
	treeStateRepositoryImpl = repository.NewTreeStateRepositoryImpl(sugaredLogger, gormDB)
	kvstateRepositoryImpl = repository.NewKVStateRepositoryImpl(sugaredLogger, gormDB)
	setStateRepositoryImpl = repository.NewSetStateRepositoryImpl(sugaredLogger, gormDB)
	actionRepositoryImpl = repository.NewActionRepositoryImpl(sugaredLogger, gormDB)
	appRepositoryImpl = repository.NewAppRepositoryImpl(sugaredLogger, gormDB)
	appSnapshotRepositoryImpl = repository.NewAppSnapshotRepositoryImpl(sugaredLogger, gormDB)

	// init service
	tssi = state.NewTreeStateServiceImpl(sugaredLogger, treeStateRepositoryImpl)
	kvssi = state.NewKVStateServiceImpl(sugaredLogger, kvstateRepositoryImpl)
	sssi = state.NewSetStateServiceImpl(sugaredLogger, setStateRepositoryImpl)
	asi = app.NewAppServiceImpl(sugaredLogger, appRepositoryImpl, kvstateRepositoryImpl, treeStateRepositoryImpl, setStateRepositoryImpl, actionRepositoryImpl)
	rsi = resource.NewResourceServiceImpl(sugaredLogger, resourceRepositoryImpl)
	return nil
}

var hub *ws.Hub

func InitHub(asi *app.AppServiceImpl,
	rsi *resource.ResourceServiceImpl,
	tssi *state.TreeStateServiceImpl,
	kvssi *state.KVStateServiceImpl,
	sssi *state.SetStateServiceImpl,
	treeStateRepositoryImpl *repository.TreeStateRepositoryImpl,
	kVStateRepositoryImpl *repository.KVStateRepositoryImpl,
	setStateRepositoryImpl *repository.SetStateRepositoryImpl,
	actionRepositoryImpl *repository.ActionRepositoryImpl,
	appRepository *repository.AppRepositoryImpl,
	appSnapshotRepositoryImpl *repository.AppSnapshotRepositoryImpl,
) {
	hub = ws.NewHub()
	hub.SetAppServiceImpl(asi)
	hub.SetResourceServiceImpl(rsi)
	hub.SetTreeStateServiceImpl(tssi)
	hub.SetKVStateServiceImpl(kvssi)
	hub.SetSetStateServiceImpl(sssi)
	hub.SetTreeStateRepositoryImpl(treeStateRepositoryImpl)
	hub.SetKVStateRepositoryImpl(kVStateRepositoryImpl)
	hub.SetSetStateRepositoryImpl(setStateRepositoryImpl)
	hub.SetActionRepositoryImpl(actionRepositoryImpl)
	hub.SetAppRepositoryImpl(appRepositoryImpl)
	hub.SetAppSnapshotRepositoryImpl(appSnapshotRepositoryImpl)
	go filter.Run(hub)
}

// ServeWebsocket handle websocket requests from the peer.
func ServeWebsocket(hub *ws.Hub, w http.ResponseWriter, r *http.Request, teamID int, appID int, clientType int) {
	// init dashbroad websocket hub

	// @todo: this CheckOrigin method for debug only, remove it for release.
	upgrader := gws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Not a web socket connection: %s \n", err)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}
	client := ws.NewClient(hub, conn, teamID, appID, clientType)
	// checkout client type
	switch clientType {
	case ws.CLIENT_TYPE_TEXT:
		client.Hub.Register <- client
	case ws.CLIENT_TYPE_BINARY:
		client.Hub.RegisterBinary <- client
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}

func main() {
	// set trial key for self-host users
	os.Setenv("ILLA_SECRET_KEY", "8xEMrWkBARcDDYQ")
	// init
	addr := flag.String("addr", "0.0.0.0:8002", "websocket server serve address")
	flag.Parse()

	// init
	initEnv()
	InitHub(asi, rsi, tssi, kvssi, sssi, treeStateRepositoryImpl, kvstateRepositoryImpl, setStateRepositoryImpl, actionRepositoryImpl, appRepositoryImpl, appSnapshotRepositoryImpl)

	// listen and serve
	r := mux.NewRouter()
	// handle /status
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	// handle ws://{ip:port}/teams/{teamID}/room/websocketConnection/dashboard
	r.HandleFunc("/teams/{teamID}/room/websocketConnection/dashboard", func(w http.ResponseWriter, r *http.Request) {
		teamID := mux.Vars(r)["teamID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		log.Printf("[Connected] /teams/%d/dashboard", teamIDInt)
		ServeWebsocket(hub, w, r, teamIDInt, ws.DASHBOARD_APP_ID, ws.CLIENT_TYPE_TEXT)
	})
	// handle ws://{ip:port}/teams/{teamID}/room/websocketConnection/apps/{appID}
	r.HandleFunc("/teams/{teamID}/room/websocketConnection/apps/{appID}", func(w http.ResponseWriter, r *http.Request) {
		teamID := mux.Vars(r)["teamID"]
		appID := mux.Vars(r)["appID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		appIDInt := idconvertor.ConvertStringToInt(appID)
		log.Printf("[Connected] /teams/%d/app/%d", teamIDInt, appIDInt)
		ServeWebsocket(hub, w, r, teamIDInt, appIDInt, ws.CLIENT_TYPE_TEXT)
	})
	// handle ws://{ip:port}/teams/{teamID}/room/binaryWebsocketConnection/apps/{appID}
	r.HandleFunc("/teams/{teamID}/room/binaryWebsocketConnection/apps/{appID}", func(w http.ResponseWriter, r *http.Request) {
		teamID := mux.Vars(r)["teamID"]
		appID := mux.Vars(r)["appID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		appIDInt := idconvertor.ConvertStringToInt(appID)
		log.Printf("[Connected] binary /teams/%d/app/%d", teamIDInt, appIDInt)
		ServeWebsocket(hub, w, r, teamIDInt, appIDInt, ws.CLIENT_TYPE_BINARY)
	})
	srv := &http.Server{
		Handler:      r,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("[START] websocket service serve on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}
