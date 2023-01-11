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

type ActionRouter interface {
	InitActionRouter(actionRouter *gin.RouterGroup)
}

type ActionRouterImpl struct {
	actionRestHandler resthandler.ActionRestHandler
}

func NewActionRouterImpl(actionRestHandler resthandler.ActionRestHandler) *ActionRouterImpl {
	return &ActionRouterImpl{actionRestHandler: actionRestHandler}
}

func (impl ActionRouterImpl) InitActionRouter(actionRouter *gin.RouterGroup) {
	actionRouter.GET("/actions", impl.actionRestHandler.FindActions)
	actionRouter.POST("/actions", impl.actionRestHandler.CreateAction)
	actionRouter.GET("/actions/:action", impl.actionRestHandler.GetAction)
	actionRouter.PUT("/actions/:action", impl.actionRestHandler.UpdateAction)
	actionRouter.DELETE("/actions/:action", impl.actionRestHandler.DeleteAction)
	actionRouter.POST("/actions/preview", impl.actionRestHandler.PreviewAction)
	actionRouter.POST("/actions/:action/run", impl.actionRestHandler.RunAction)
}
