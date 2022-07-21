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
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/internal/util"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/db"
	"github.com/illa-family/builder-backend/pkg/resource"
	"github.com/illa-family/builder-backend/pkg/state"

	ws "github.com/illa-family/builder-backend/internal/websocket"
	ws_server "github.com/illa-family/builder-backend/pkg/websocket-server"
)

// websocket client hub

var tssi *state.TreeStateServiceImpl
var kvssi *state.KVStateServiceImpl
var sssi *state.SetStateServiceImpl
var asi *app.AppServiceImpl
var rsi *resource.ResourceServiceImpl

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
	treestateRepositoryImpl := repository.NewTreeStateRepositoryImpl(sugaredLogger, gormDB)
	kvstateRepositoryImpl := repository.NewKVStateRepositoryImpl(sugaredLogger, gormDB)
	setstateRepositoryImpl := repository.NewSetStateRepositoryImpl(sugaredLogger, gormDB)
	appRepositoryImpl := repository.NewAppRepositoryImpl(sugaredLogger, gormDB)
	resourceRepositoryImpl := repository.NewResourceRepositoryImpl(sugaredLogger, gormDB)
	userRepositoryImpl := repository.NewUserRepositoryImpl(gormDB, sugaredLogger)
	actionRepositoryImpl := repository.NewActionRepositoryImpl(sugaredLogger, gormDB)
	// init service
	tssi = state.NewTreeStateServiceImpl(sugaredLogger, treestateRepositoryImpl)
	kvssi = state.NewKVStateServiceImpl(sugaredLogger, kvstateRepositoryImpl)
	sssi = state.NewSetStateServiceImpl(sugaredLogger, setstateRepositoryImpl)
	asi = app.NewAppServiceImpl(sugaredLogger, appRepositoryImpl, userRepositoryImpl, kvstateRepositoryImpl, treestateRepositoryImpl, actionRepositoryImpl)
	rsi = resource.NewResourceServiceImpl(sugaredLogger, resourceRepositoryImpl)
	return nil
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "websocket server serve address")
	flag.Parse()

	// init
	initEnv()
	ws_server.InitHub(asi, rsi, tssi, kvssi, sssi)

	// listen and serve
	r := mux.NewRouter()
	// handle /status
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	// handle ws://{ip:port}/room/{instanceID}/dashboard
	r.HandleFunc("/room/{instanceID}/dashboard", func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instanceID"]
		log.Printf("[Connected] /room/%s/dashboard", instanceID)
		ws_server.ServeWebsocket(w, r, instanceID, ws.DEAULT_APP_ID)
	})
	// handle ws://{ip:port}/room/{instanceID}/app/{appID}
	r.HandleFunc("/room/{instanceID}/app/{appID}", func(w http.ResponseWriter, r *http.Request) {
		instanceID := mux.Vars(r)["instanceID"]
		appID, err := strconv.Atoi(mux.Vars(r)["appID"])
		if err != nil {
			appID = ws.DEAULT_APP_ID
		}
		log.Printf("[Connected] /room/%s/app/%d", instanceID, appID)
		ws_server.ServeWebsocket(w, r, instanceID, appID)
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
