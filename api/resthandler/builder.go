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
	"net/http"

	ac "github.com/illacloud/illa-builder-backend/internal/accesscontrol"
	"github.com/illacloud/illa-builder-backend/pkg/builder"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BuilderRestHandler interface {
	GetTeamBuilderDesc(c *gin.Context)
}

type BuilderRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	builderService builder.BuilderService
	AttributeGroup *ac.AttributeGroup
}

func NewBuilderRestHandlerImpl(logger *zap.SugaredLogger, builderService builder.BuilderService, attrg *ac.AttributeGroup) *BuilderRestHandlerImpl {
	return &BuilderRestHandlerImpl{
		logger:         logger,
		builderService: builderService,
		AttributeGroup: attrg,
	}
}

func (impl BuilderRestHandlerImpl) GetTeamBuilderDesc(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	impl.AttributeGroup.Init()
	impl.AttributeGroup.SetTeamID(teamID)
	impl.AttributeGroup.SetUserAuthToken(userAuthToken)
	impl.AttributeGroup.SetUnitType(ac.UNIT_TYPE_BUILDER_DASHBOARD)
	impl.AttributeGroup.SetUnitID(ac.DEFAULT_UNIT_ID)
	canAccess, errInCheckAttr := impl.AttributeGroup.CanAccess(ac.ACTION_ACCESS_VIEW)
	if errInCheckAttr != nil {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	ret, err := impl.builderService.GetTeamBuilderDesc(teamID)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_BUILDER_DESCRIPTION, "get builder description error: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, ret)
}
