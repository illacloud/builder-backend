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
	"github.com/gin-gonic/gin"
	"github.com/illa-family/builder-backend/api/resthandler"
)

type AppRouter interface {
	InitAppRouter(actionRouter *gin.RouterGroup)
}

type AppRouterImpl struct {
	appRestHandler resthandler.AppRestHandler
}

func NewAppRouterImpl(appRestHandler resthandler.AppRestHandler) *AppRouterImpl {
	return &AppRouterImpl{appRestHandler: appRestHandler}
}

func (impl AppRouterImpl) InitAppRouter(appRouter *gin.RouterGroup) {
	appRouter.POST("/apps", impl.appRestHandler.CreateApp)
	appRouter.DELETE("/apps/:id", impl.appRestHandler.DeleteApp)
	appRouter.PUT("/apps/:id", impl.appRestHandler.RenameApp)
	appRouter.GET("/apps", impl.appRestHandler.GetAllApp)
}
