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

	"go.uber.org/zap"
)

const DEFAULT_SERVER_ADDRESS = "localhost"
const DEFAULT_WEBSOCKET_PORT = "8000"
const PROTOCOL_WEBSOCKET = "ws"
const PROTOCOL_WEBSOCKET_OVER_TLS = "wss"
const DASHBOARD_WS_URL = "%s://%s:%s/builder/teams/%s/dashboard"
const ROOM_WS_URL = "%s://%s:%s/builder/teams/%s/app/%d"

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

func (impl *RoomServiceImpl) GetDashboardConn(teamID int) (WSURLResponse, error) {
	var r WSURLResponse
	r.WSURL = fmt.Sprintf(DASHBOARD_WS_URL, getProtocol(), getServerAddress(), getWebSocketPort(), teamID)
	return r, nil
}

func (impl *RoomServiceImpl) GetAppRoomConn(teamID int, roomID int) (WSURLResponse, error) {
	var r WSURLResponse
	r.WSURL = fmt.Sprintf(ROOM_WS_URL, getProtocol(), getServerAddress(), getWebSocketPort(), teamID, roomID)
	return r, nil
}
