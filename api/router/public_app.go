// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by publicApplicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"github.com/illacloud/builder-backend/api/resthandler"

	"github.com/gin-gonic/gin"
)

type PublicAppRouter interface {
	InitPublicAppRouter(actionRouter *gin.RouterGroup)
}

type PublicAppRouterImpl struct {
	publicAppRestHandler resthandler.PublicAppRestHandler
}

func NewPublicAppRouterImpl(publicAppRestHandler resthandler.PublicAppRestHandler) *PublicAppRouterImpl {
	return &PublicAppRouterImpl{publicAppRestHandler: publicAppRestHandler}
}

func (impl PublicAppRouterImpl) InitPublicAppRouter(publicAppRouter *gin.RouterGroup) {
	publicAppRouter.GET(":appID/versions/:version", impl.publicAppRestHandler.GetMegaData)
}
