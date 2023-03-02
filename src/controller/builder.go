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

package controller

import (
	"net/http"
	"time"

	ac "github.com/illacloud/illa-builder-backend/src/accesscontrol"
	"github.com/illacloud/illa-builder-backend/src/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BuilderController interface {
	GetTeamBuilderDesc(c *gin.Context)
}

type BuilderControllerImpl struct {
	logger         *zap.SugaredLogger
	Storage        *model.Storage
	AttributeGroup *ac.AttributeGroup
}

func NewBuilderControllerImpl(logger *zap.SugaredLogger, attrg *ac.AttributeGroup) *BuilderControllerImpl {
	return &BuilderControllerImpl{
		logger:         logger,
		AttributeGroup: attrg,
	}
}

func (impl BuilderControllerImpl) GetTeamBuilderDesc(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate attribute
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

	// fetch team app num
	appNum, errInFetchAppNum := impl.Storage.AppStorage.CountAPPByTeamID(teamID)
	if errInFetchAppNum != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_BUILDER_DESCRIPTION, "fetch app num failed: "+errInFetchAppNum.Error())
		return
	}
	// fetch team resource num
	resourceNum, errInFetchResourceNum := impl.Storage.AppStorage.CountResourceByTeamID(teamID)
	if errInFetchResourceNum != nil {
		return nil, errInFetchResourceNum
	}
	actionNum, errInFetchAactionNum := impl.actionRepository.CountActionByTeamID(teamID)
	if errInFetchAactionNum != nil {
		return nil, errInFetchAactionNum
	}
	appLastModifyedAt, errInFetchAppModifyTime := impl.appRepository.RetrieveAppLastModifiedTime(teamID)
	resourceLastModifyedAt, errInFetchResourceModifyTime := impl.resourceRepository.RetrieveResourceLastModifiedTime(teamID)

	// compare time
	var lastModifiedAt time.Time
	if errInFetchAppModifyTime == nil && errInFetchResourceModifyTime == nil {
		if appLastModifyedAt.Before(resourceLastModifyedAt) {
			lastModifiedAt = resourceLastModifyedAt
		} else {
			lastModifiedAt = appLastModifyedAt
		}
	} else if errInFetchResourceModifyTime != nil {
		lastModifiedAt = appLastModifyedAt
	} else if errInFetchAppModifyTime != nil {
		lastModifiedAt = resourceLastModifyedAt
	}

	if errInFetchAppModifyTime != nil && errInFetchResourceModifyTime != nil {
		return NewEmptyBuilderDescResponse(resourceNum, resourceNum, actionNum), nil
	}

	ret := NewGetBuilderDescResponse(appNum, resourceNum, actionNum, lastModifiedAt)
	return ret, nil

	// fetch data
	ret, err := impl.builderService.GetTeamBuilderDesc(teamID)
	if err != nil {
		FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_BUILDER_DESCRIPTION, "get builder description error: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, ret)
}
