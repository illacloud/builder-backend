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
	"time"

	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) GetTeamBuilderDesc(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_BUILDER_DASHBOARD,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_ACCESS_VIEW,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canAccess {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch data
	appNum, _ := controller.Storage.AppStorage.CountAPPByTeamID(teamID)
	resourceNum, _ := controller.Storage.ResourceStorage.CountResourceByTeamID(teamID)
	actionNum, _ := controller.Storage.ActionStorage.CountActionByTeamID(teamID)
	appLastModifyedAt, errInFetchAppModifyTime := controller.Storage.AppStorage.RetrieveAppLastModifiedTime(teamID)
	resourceLastModifyedAt, errInFetchResourceModifyTime := controller.Storage.ResourceStorage.RetrieveResourceLastModifiedTime(teamID)

	// team have no app and no resource
	if errInFetchAppModifyTime != nil && errInFetchResourceModifyTime != nil {
		feed := response.NewEmptyBuilderDescResponse(resourceNum, resourceNum, actionNum)
		controller.FeedbackOK(c, feed)
		return
	}

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

	// feedback
	feed := response.NewGetBuilderDescResponse(appNum, resourceNum, actionNum, lastModifiedAt)
	controller.FeedbackOK(c, feed)
	return
}
