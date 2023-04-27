// Copyright 2023 Illa Soft, Inc.
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
	"github.com/illacloud/builder-backend/api/resthandler"
)

type OAuth2Router interface {
	InitOAuth2Router(oauth2Router *gin.RouterGroup)
}

type OAuth2RouterImpl struct {
	OAuth2RestHandler resthandler.OAuth2RestHandler
}

func NewOAuth2RouterImpl(OAuth2RestHandler resthandler.OAuth2RestHandler) *OAuth2RouterImpl {
	return &OAuth2RouterImpl{OAuth2RestHandler: OAuth2RestHandler}
}

func (impl OAuth2RouterImpl) InitOAuth2Router(oauth2Router *gin.RouterGroup) {
	oauth2Router.GET("/authorize", impl.OAuth2RestHandler.GoogleOAuth2)
}
