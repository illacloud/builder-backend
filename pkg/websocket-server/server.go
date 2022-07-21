package websocket_server

import (
	"log"
	"net/http"

	gws "github.com/gorilla/websocket"
	ws "github.com/illa-family/builder-backend/internal/websocket"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/resource"
	"github.com/illa-family/builder-backend/pkg/state"
)

var dashboardHub *ws.Hub
var appHub *ws.Hub

func InitHub(asi *app.AppServiceImpl, rsi *resource.ResourceServiceImpl, tssi *state.TreeStateServiceImpl, kvssi *state.KVStateServiceImpl, sssi *state.SetStateServiceImpl) {
	dashboardHub = ws.NewHub()
	dashboardHub.SetAppServiceImpl(asi)
	go dashboardHub.Run()

	// init APP websocket hub
	appHub = ws.NewHub()
	appHub.SetResourceServiceImpl(rsi)
	appHub.SetTreeStateServiceImpl(tssi)
	appHub.SetKVStateServiceImpl(kvssi)
	appHub.SetSetStateServiceImpl(sssi)
	go appHub.Run()
}

// ServeWebsocket handle websocket requests from the peer.
func ServeWebsocket(hub *ws.Hub, w http.ResponseWriter, r *http.Request, instanceID string, appID int) {
	// init dashbroad websocket hub

	// @todo: this CheckOrigin method for debug only, remove it for release.
	var upgrader = gws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Not a web socket connection: %s \n", err)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}
	client := ws.NewClient(hub, conn, instanceID, appID)
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
