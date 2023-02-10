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

package router

import (
	"github.com/illacloud/builder-backend/api/resthandler"

	"github.com/gin-gonic/gin"
)

type RoomRouter interface {
	InitRoomRouter(roomRouter *gin.RouterGroup)
}

type RoomRouterImpl struct {
	RoomRestHandler resthandler.RoomRestHandler
}

func NewRoomRouterImpl(RoomRestHandler resthandler.RoomRestHandler) *RoomRouterImpl {
	return &RoomRouterImpl{RoomRestHandler: RoomRestHandler}
}

func (impl RoomRouterImpl) InitRoomRouter(roomRouter *gin.RouterGroup) {
	roomRouter.GET("/:instanceID/dashboard", impl.RoomRestHandler.GetDashboardRoomConn)
	roomRouter.GET("/:instanceID/app/:roomID", impl.RoomRestHandler.GetAppRoomConn)
}
