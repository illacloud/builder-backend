// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by publicActionlicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"github.com/illacloud/builder-backend/api/resthandler"

	"github.com/gin-gonic/gin"
)

type PublicActionRouter interface {
	InitPublicActionRouter(actionRouter *gin.RouterGroup)
}

type PublicActionRouterImpl struct {
	publicActionRestHandler resthandler.PublicActionRestHandler
}

func NewPublicActionRouterImpl(publicActionRestHandler resthandler.PublicActionRestHandler) *PublicActionRouterImpl {
	return &PublicActionRouterImpl{publicActionRestHandler: publicActionRestHandler}
}

func (impl PublicActionRouterImpl) InitPublicActionRouter(publicActionRouter *gin.RouterGroup) {
	publicActionRouter.POST("/:actionID/run", impl.publicActionRestHandler.RunAction)
}
