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

type UserRouter interface {
	InitAuthRouter(authRouter *gin.RouterGroup)
	InitUserRouter(userRouter *gin.RouterGroup)
}

type UserRouterImpl struct {
	userRestHandler resthandler.UserRestHandler
}

func NewUserRouterImpl(userRestHandler resthandler.UserRestHandler) *UserRouterImpl {
	return &UserRouterImpl{userRestHandler: userRestHandler}
}

func (impl UserRouterImpl) InitAuthRouter(authRouter *gin.RouterGroup) {
	authRouter.POST("/verification", impl.userRestHandler.GetVerificationCode)
	authRouter.POST("/signup", impl.userRestHandler.SignUp)
	authRouter.POST("/signin", impl.userRestHandler.SignIn)
	authRouter.POST("/forgetPassword", impl.userRestHandler.ForgetPassword)
}

func (impl UserRouterImpl) InitUserRouter(userRouter *gin.RouterGroup) {
	userRouter.PATCH("/password", impl.userRestHandler.UpdatePassword)
	userRouter.PATCH("/nickname", impl.userRestHandler.UpdateUsername)
	userRouter.PATCH("/language", impl.userRestHandler.UpdateLanguage)
	userRouter.GET("", impl.userRestHandler.GetUserInfo)
}
