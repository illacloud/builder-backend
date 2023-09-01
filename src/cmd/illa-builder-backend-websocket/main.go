package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	gws "github.com/gorilla/websocket"
	"github.com/illacloud/builder-backend/src/driver/postgres"
	"github.com/illacloud/builder-backend/src/storage"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/builderoperation"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/logger"
	"github.com/illacloud/builder-backend/src/utils/supervisor"
	"github.com/illacloud/builder-backend/src/websocket"
	filter "github.com/illacloud/builder-backend/src/websocket-filter"
	"go.uber.org/zap"
)

var hub *websocket.Hub

func InitStorage(globalConfig *config.Config, logger *zap.SugaredLogger) *storage.Storage {
	postgresDriver, err := postgres.NewPostgresConnectionByGlobalConfig(globalConfig, logger)
	if err != nil {
		logger.Errorw("Error in startup, storage init failed.")
	}
	return storage.NewStorage(postgresDriver, logger)
}

func InitHub(s *storage.Storage) {
	// init attribute group
	attrg, errInNewAttributeGroup := accesscontrol.NewRawAttributeGroup()
	if errInNewAttributeGroup != nil {
		log.Fatalf("Error in startup, attribute group init failed.")
		return
	}

	// new hub
	hub = websocket.NewHub(s, attrg)
	go filter.Run(hub)
}

// ServeWebsocket handle websocket requests from the peer.
func ServeWebsocket(hub *websocket.Hub, w http.ResponseWriter, r *http.Request, teamID int, appID int, clientType int) {
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
	client := websocket.NewClient(hub, conn, teamID, appID, clientType)
	// checkout client type
	switch clientType {
	case websocket.CLIENT_TYPE_TEXT:
		client.Hub.Register <- client
	case websocket.CLIENT_TYPE_BINARY:
		client.Hub.RegisterBinary <- client
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}

func main() {
	// init
	conf := config.GetInstance()
	sugaredLogger := logger.NewSugardLogger()

	storage := InitStorage(conf, sugaredLogger)
	InitHub(storage)

	// listen and serve
	r := mux.NewRouter()

	// handle /status
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// handle /status
	r.HandleFunc("/api/v1/teams/{teamID}/apps/{appID}/recoverSnapshot", func(w http.ResponseWriter, r *http.Request) {
		// set cors
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, "+
			"Access-Control-Allow-Headers, Authorization, Cache-Control, Content-Language, Content-Type, illa-token")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		w.Header().Add("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// get teamID & appID
		teamID := mux.Vars(r)["teamID"]
		appID := mux.Vars(r)["appID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		appIDInt := idconvertor.ConvertStringToInt(appID)

		// check user authorization
		authorizationToken := r.Header.Get("Authorization")
		supervisor := supervisor.NewSupervisor()
		validated, errInValidate := supervisor.ValidateUserAccount(authorizationToken)
		if errInValidate != nil {
			return
		}
		if !validated {
			return
		}

		// check if user have access permission to target team and app
		attributeGroup, _ := accesscontrol.NewRawAttributeGroup()
		canManage, errInCheckAttr := attributeGroup.CanManage(
			teamIDInt,
			authorizationToken,
			accesscontrol.UNIT_TYPE_APP,
			appIDInt,
			accesscontrol.ACTION_MANAGE_EDIT_APP,
		)
		if errInCheckAttr != nil {
			return
		}
		if !canManage {
			return
		}

		// ok, broadcast refresh message to room all client
		serverSideClientID := websocket.GetMessageClientIDForWebsocketServer()
		message, errInNewWebSocketMessage := websocket.NewEmptyMessage(appIDInt, serverSideClientID, builderoperation.SIGNAL_FORCE_REFRESH, builderoperation.TARGET_WINDOW, true)
		if errInNewWebSocketMessage != nil {
			json.NewEncoder(w).Encode(map[string]bool{"ok": true})
			return
		}
		hub.SendFeedbackToTargetRoomAllClients(websocket.ERROR_FORCE_REFRESH_WINDOW, message, teamIDInt, appIDInt)

		// done
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// handle ws://{ip:port}/teams/{teamID}/room/websocketConnection/dashboard
	r.HandleFunc("/teams/{teamID}/room/websocketConnection/dashboard", func(w http.ResponseWriter, r *http.Request) {
		teamID := mux.Vars(r)["teamID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		log.Printf("[Connected] /teams/%d/dashboard", teamIDInt)
		ServeWebsocket(hub, w, r, teamIDInt, websocket.DASHBOARD_APP_ID, websocket.CLIENT_TYPE_TEXT)
	})

	// handle ws://{ip:port}/teams/{teamID}/room/websocketConnection/apps/{appID}
	r.HandleFunc("/teams/{teamID}/room/websocketConnection/apps/{appID}", func(w http.ResponseWriter, r *http.Request) {
		teamID := mux.Vars(r)["teamID"]
		appID := mux.Vars(r)["appID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		appIDInt := idconvertor.ConvertStringToInt(appID)
		log.Printf("[Connected] /teams/%d/app/%d", teamIDInt, appIDInt)
		ServeWebsocket(hub, w, r, teamIDInt, appIDInt, websocket.CLIENT_TYPE_TEXT)
	})

	// handle ws://{ip:port}/teams/{teamID}/room/binaryWebsocketConnection/apps/{appID}
	r.HandleFunc("/teams/{teamID}/room/binaryWebsocketConnection/apps/{appID}", func(w http.ResponseWriter, r *http.Request) {
		teamID := mux.Vars(r)["teamID"]
		appID := mux.Vars(r)["appID"]
		teamIDInt := idconvertor.ConvertStringToInt(teamID)
		appIDInt := idconvertor.ConvertStringToInt(appID)
		log.Printf("[Connected] binary /teams/%d/app/%d", teamIDInt, appIDInt)
		ServeWebsocket(hub, w, r, teamIDInt, appIDInt, websocket.CLIENT_TYPE_BINARY)
	})

	server := &http.Server{
		Handler:      r,
		Addr:         conf.GetWebScoketServerListenAddress(),
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	log.Printf("[START] illa-builder-backend-websocket service serve on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
