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

package wireset

import (
	"github.com/google/wire"

	"github.com/illacloud/builder-backend/api/resthandler"
	"github.com/illacloud/builder-backend/api/router"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/user"
)

var UserWireSet = wire.NewSet(
	repository.NewUserRepositoryImpl,
	wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),
	user.NewUserServiceImpl,
	wire.Bind(new(user.UserService), new(*user.UserServiceImpl)),
	user.NewAuthenticatorImpl,
	wire.Bind(new(user.Authenticator), new(*user.AuthenticatorImpl)),
	resthandler.NewUserRestHandlerImpl,
	wire.Bind(new(resthandler.UserRestHandler), new(*resthandler.UserRestHandlerImpl)),
	router.NewUserRouterImpl,
	wire.Bind(new(router.UserRouter), new(*router.UserRouterImpl)),
)
