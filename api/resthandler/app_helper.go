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

package resthandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	"github.com/illacloud/builder-backend/internal/auditlogger"
	"github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/action"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func (impl AppRestHandlerImpl) SnapshotTreeState(c *gin.Context, teamID int, appID int, appMainLineVersion int) {
	// get edit version K-V state from database
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(teamID, appID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	indexIDMap := map[int]int{}
	releaseIDMap := map[int]int{}

	// set version as mainline version
	for serial, _ := range treestates {
		indexIDMap[serial] = treestates[serial].ID
		treestates[serial].ID = 0
		treestates[serial].UID = uuid.New()
		treestates[serial].Version = appMainLineVersion
	}

	// and put them to the database as duplicate
	for i, treestate := range treestates {
		id, err := impl.treestateRepository.Create(treestate)
		if err != nil {
			return err
		}
		oldID := indexIDMap[i]
		releaseIDMap[oldID] = id
	}

	for _, treestate := range treestates {
		treestate.ChildrenNodeRefIDs = convertLink(treestate.ChildrenNodeRefIDs, releaseIDMap)
		treestate.ParentNodeRefID = releaseIDMap[treestate.ParentNodeRefID]
		if err := impl.treestateRepository.Update(treestate); err != nil {
			return err
		}
	}

	return nil
}
