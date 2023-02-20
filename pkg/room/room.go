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

package room

import (
	"fmt"
	"os"

	"github.com/illacloud/builder-backend/internal/idconvertor"
	"go.uber.org/zap"
)

const DEFAULT_SERVER_ADDRESS = "localhost"
const DEFAULT_WEBSOCKET_PORT = "8000"
const PROTOCOL_WEBSOCKET = "ws"
const PROTOCOL_WEBSOCKET_OVER_TLS = "wss"
const ILLA_DEPLOY_MODE_SELF_HOST = "self-host"
const ILLA_DEPLOY_MODE_CLOUD = "cloud"
const DASHBOARD_WS_URL = "%s://%s:%s/teams/%s/room/websocketConnection/dashboard"
const ROOM_WS_URL = "%s://%s:%s/teams/%s/room/websocketConnection/apps/%s"
const SELF_HOST_DASHBOARD_WS_URL = "/builder-ws/teams/%s/room/websocketConnection/dashboard"
const SELF_HOST_ROOM_WS_URL = "/builder-ws/teams/%s/room/websocketConnection/apps/%s"

type RoomService interface {
	GetDashboardConn(teamID int) (WSURLResponse, error)
	GetAppRoomConn(teamID int, roomID int) (WSURLResponse, error)
}

type RoomServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewRoomServiceImpl(logger *zap.SugaredLogger) *RoomServiceImpl {
	return &RoomServiceImpl{
		logger: logger,
	}
}

type WSURLResponse struct {
	WSURL string `json:"wsURL"`
}

func getProtocol() string {
	wssEnabled := os.Getenv("WSS_ENABLED")
	if wssEnabled == "true" {
		return PROTOCOL_WEBSOCKET_OVER_TLS
	}
	return PROTOCOL_WEBSOCKET
}

func getServerAddress() string {
	serverAddress := os.Getenv("WEBSOCKET_SERVER_ADDRESS")
	if len(serverAddress) == 0 || serverAddress == "" {
		return DEFAULT_SERVER_ADDRESS
	}
	return serverAddress
}

func getWebSocketPort() string {
	webSockerPort := os.Getenv("WEBSOCKET_PORT")
	if len(webSockerPort) == 0 || webSockerPort == "" {
		return DEFAULT_WEBSOCKET_PORT
	}
	return webSockerPort
}

// self-host mode by default
func isSelfHostDeployMode() bool {
	mode := os.Getenv("ILLA_DEPLOY_MODE")
	if len(mode) == 0 || mode == "" || mode == ILLA_DEPLOY_MODE_SELF_HOST {
		return true
	}
	if mode == ILLA_DEPLOY_MODE_CLOUD {
		return false
	}
	return true
}

func (impl *RoomServiceImpl) GetDashboardConn(teamID int) (WSURLResponse, error) {
	var r WSURLResponse
	if isSelfHostDeployMode() {
		r.WSURL = fmt.Sprintf(DASHBOARD_WS_URL, getProtocol(), getServerAddress(), getWebSocketPort(), idconvertor.ConvertIntToString(teamID))
	} else {
		r.WSURL = fmt.Sprintf(SELF_HOST_DASHBOARD_WS_URL, idconvertor.ConvertIntToString(teamID))
	}
	return r, nil
}

func (impl *RoomServiceImpl) GetAppRoomConn(teamID int, roomID int) (WSURLResponse, error) {
	var r WSURLResponse
	if isSelfHostDeployMode() {
		r.WSURL = fmt.Sprintf(ROOM_WS_URL, getProtocol(), getServerAddress(), getWebSocketPort(), idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	} else {
		r.WSURL = fmt.Sprintf(SELF_HOST_ROOM_WS_URL, idconvertor.ConvertIntToString(teamID), idconvertor.ConvertIntToString(roomID))
	}
	return r, nil
}
