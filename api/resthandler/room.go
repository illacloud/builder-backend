// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by Roomlicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resthandler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/illa-family/builder-backend/pkg/room"
	"go.uber.org/zap"
)

type RoomRequest struct {
	Name string `json:"RoomName" validate:"required"`
}

type RoomRestHandler interface {
	GetDashboardRoomConn(c *gin.Context)
	GetAppRoomConn(c *gin.Context)
}

type RoomRestHandlerImpl struct {
	logger      *zap.SugaredLogger
	RoomService room.RoomService
}

func NewRoomRestHandlerImpl(logger *zap.SugaredLogger, RoomService room.RoomService) *RoomRestHandlerImpl {
	return &RoomRestHandlerImpl{
		logger:      logger,
		RoomService: RoomService,
	}
}

func (impl RoomRestHandlerImpl) GetDashboardRoomConn(c *gin.Context) {
	// Get User from auth middleware
	instanceID, okGet := c.Get("instanceID")
	log.Printf("[DUMP] instanceID: %v\n", instanceID)

	iid, okReflect := instanceID.(string)
	if !(okGet && okReflect) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	roomData, _ := impl.RoomService.GetDashboardConn(iid)

	c.JSON(http.StatusOK, roomData)
}

func (impl RoomRestHandlerImpl) GetAppRoomConn(c *gin.Context) {
	// Get User from auth middleware
	instanceID, ok1 := c.Get("instanceID")
	roomID, ok2 := c.Get("roomID")
	iid, okReflect1 := instanceID.(string)
	rid, okReflect2 := roomID.(int)
	if !(ok1 && ok2 && okReflect1 && okReflect2) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	roomData, _ := impl.RoomService.GetAppRoomConn(iid, rid)

	c.JSON(http.StatusOK, roomData)
}
