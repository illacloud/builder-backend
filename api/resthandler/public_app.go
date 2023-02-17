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

package resthandler

import (
	ac "github.com/illacloud/builder-backend/internal/accesscontrol"
	dc "github.com/illacloud/builder-backend/internal/datacontrol"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"github.com/illacloud/builder-backend/pkg/state"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PublicAppRestHandler interface {
	GetMegaData(c *gin.Context)
}

type PublicAppRestHandlerImpl struct {
	logger           *zap.SugaredLogger
	appService       app.AppService
	AttributeGroup   *ac.AttributeGroup
	treeStateService state.TreeStateService
}

func NewPublicAppRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppService, attrg *ac.AttributeGroup, treeStateService state.TreeStateService) *PublicAppRestHandlerImpl {
	return &PublicAppRestHandlerImpl{
		logger:           logger,
		appService:       appService,
		AttributeGroup:   attrg,
		treeStateService: treeStateService,
	}
}

func (impl PublicAppRestHandlerImpl) GetMegaData(c *gin.Context) {
	// fetch needed param
	teamIdentifier, errInGetTeamID := GetStringParamFromRequest(c, PARAM_TEAM_IDENTIFIER)
	publicAppID, errInGetAPPID := GetMagicIntParamFromRequest(c, PARAM_APP_ID)
	version, errInGetVersion := GetIntParamFromRequest(c, PARAM_VERSION)
	if errInGetTeamID != nil || errInGetAPPID != nil || errInGetVersion != nil {
		return
	}

	// check version, the version must be repository.APP_AUTO_RELEASE_VERSION
	if version != repository.APP_AUTO_RELEASE_VERSION {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you only can access release version of app.")
		return
	}

	// get team id by team teamIdentifier
	team, errInGetTeamInfo := dc.GetTeamInfoByIdentifier(teamIdentifier)
	if errInGetTeamInfo != nil {
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_TEAM, "get target team by identifier error: "+errInGetTeamInfo.Error())
		return
	}
	teamID := team.GetID()

	// check if app is public app
	if !impl.appService.IsPublicApp(teamID, publicAppID) {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this app.")
		return
	}

	// Fetch Mega data via `publicApp` and `version`
	res, err := impl.appService.GetMegaData(teamID, publicAppID, version)
	if err != nil {
		if err.Error() == "content not found" {
			FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get publicApp mega data error: "+err.Error())
			return
		}
		FeedbackInternalServerError(c, ERROR_FLAG_CAN_NOT_GET_APP, "get publicApp mega data error: "+err.Error())
		return
	}

	// feedback
	FeedbackOK(c, res)
	return
}
