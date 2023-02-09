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
	"github.com/illacloud/builder-backend/api/resthandler"
	"github.com/illacloud/builder-backend/api/router"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	"github.com/google/wire"
)

var AppWireSet = wire.NewSet(
	repository.NewKVStateRepositoryImpl,
	wire.Bind(new(repository.KVStateRepository), new(*repository.KVStateRepositoryImpl)),
	repository.NewTreeStateRepositoryImpl,
	wire.Bind(new(repository.TreeStateRepository), new(*repository.TreeStateRepositoryImpl)),
	repository.NewSetStateRepositoryImpl,
	wire.Bind(new(repository.SetStateRepository), new(*repository.SetStateRepositoryImpl)),
	repository.NewAppRepositoryImpl,
	wire.Bind(new(repository.AppRepository), new(*repository.AppRepositoryImpl)),
	app.NewAppServiceImpl,
	wire.Bind(new(app.AppService), new(*app.AppServiceImpl)),
	state.NewTreeStateServiceImpl,
	wire.Bind(new(state.TreeStateService), new(*state.TreeStateServiceImpl)),
	resthandler.NewAppRestHandlerImpl,
	wire.Bind(new(resthandler.AppRestHandler), new(*resthandler.AppRestHandlerImpl)),
	router.NewAppRouterImpl,
	wire.Bind(new(router.AppRouter), new(*router.AppRouterImpl)),
)
