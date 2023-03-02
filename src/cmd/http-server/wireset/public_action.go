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
	"github.com/illacloud/illa-builder-backend/api/resthandler"
	"github.com/illacloud/illa-builder-backend/api/router"
	"github.com/illacloud/illa-builder-backend/internal/repository"
	"github.com/illacloud/illa-builder-backend/pkg/action"

	"github.com/google/wire"
)

var PublicActionWireSet = wire.NewSet(
	repository.NewPublicActionRepositoryImpl,
	wire.Bind(new(repository.PublicActionRepository), new(*repository.PublicActionRepositoryImpl)),
	action.NewPublicActionServiceImpl,
	wire.Bind(new(action.PublicActionService), new(*action.PublicActionServiceImpl)),
	resthandler.NewPublicActionRestHandlerImpl,
	wire.Bind(new(resthandler.PublicActionRestHandler), new(*resthandler.PublicActionRestHandlerImpl)),
	router.NewPublicActionRouterImpl,
	wire.Bind(new(router.PublicActionRouter), new(*router.PublicActionRouterImpl)),
)
