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
	"github.com/illacloud/builder-backend/pkg/resource"

	"github.com/google/wire"
)

var ResourceWireSet = wire.NewSet(
	repository.NewResourceRepositoryImpl,
	wire.Bind(new(repository.ResourceRepository), new(*repository.ResourceRepositoryImpl)),
	resource.NewResourceServiceImpl,
	wire.Bind(new(resource.ResourceService), new(*resource.ResourceServiceImpl)),
	resthandler.NewResourceRestHandlerImpl,
	wire.Bind(new(resthandler.ResourceRestHandler), new(*resthandler.ResourceRestHandlerImpl)),
	router.NewResourceRouterImpl,
	wire.Bind(new(router.ResourceRouter), new(*router.ResourceRouterImpl)),
)
