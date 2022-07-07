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
	"time"

	"github.com/gorilla/mux"
	"github.com/illa-family/builder-backend/internal/websocket"
)

// websocket client hub
var dashboardHub *websocket.Hub
var appHub *websocket.Hub

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "websocket server serve address")
	flag.Parse()

	// start websocket hub
	dashboardHub = websocket.NewHub()
	appHub = websocket.NewHub()
	go dashboardHub.Run()
	go appHub.Run()

	// listen and serve
	r := mux.NewRouter()
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	r.HandleFunc("/room/dashboard/{roomId}", func(w http.ResponseWriter, r *http.Request) {
		roomId := mux.Vars(r)["roomId"]
		log.Printf("[Connected] /room/dashboard/%s", roomId)
		websocket.ServeWebsocket(dashboardHub, w, r, roomId)
	})
	r.HandleFunc("/room/app/{roomId}", func(w http.ResponseWriter, r *http.Request) {
		roomId := mux.Vars(r)["roomId"]
		websocket.ServeWebsocket(appHub, w, r, roomId)
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
