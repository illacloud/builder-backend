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
	"strconv"

	"github.com/illacloud/builder-backend/pkg/room"

	"github.com/gin-gonic/gin"
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
	instanceID := c.Param("instanceID")
	log.Printf("[DUMP] instanceID: %v\n", instanceID)

	roomData, _ := impl.RoomService.GetDashboardConn(instanceID)

	c.JSON(http.StatusOK, roomData)
}

func (impl RoomRestHandlerImpl) GetAppRoomConn(c *gin.Context) {
	// Get User from auth middleware
	instanceID := c.Param("instanceID")
	roomID := c.Param("roomID")

	rid, err := strconv.Atoi(roomID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode":    401,
			"errorMessage": "unauthorized",
		})
		return
	}

	roomData, _ := impl.RoomService.GetAppRoomConn(instanceID, rid)

	c.JSON(http.StatusOK, roomData)
}
