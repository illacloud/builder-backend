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

	"go.uber.org/zap"
)

const DASHBOARD_WS_URL = "wss://localhost:8000/room/%s/dashboard"
const ROOM_WS_URL = "wss://localhost/room/%s/app/%d"

type RoomService interface {
	GetDashboardConn(instanceID string) (WSURLResponse, error)
	GetAppRoomConn(instanceID string, roomID int) (WSURLResponse, error)
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

func (impl *RoomServiceImpl) GetDashboardConn(instanceID string) (WSURLResponse, error) {
	var r WSURLResponse
	r.WSURL = fmt.Sprintf(DASHBOARD_WS_URL, instanceID)
	return r, nil
}

func (impl *RoomServiceImpl) GetAppRoomConn(instanceID string, roomID int) (WSURLResponse, error) {
	var r WSURLResponse
	r.WSURL = fmt.Sprintf(ROOM_WS_URL, instanceID, roomID)
	return r, nil
}
