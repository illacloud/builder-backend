package controller

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
)

func (controller *Controller) SetActionTutorialLink(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	actionID, errInGetActionID := controller.GetMagicIntParamFromRequest(c, PARAM_ACTION_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetActionID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_ACTION,
		actionID,
		accesscontrol.ACTION_MANAGE_EDIT_ACTION,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// fetch payload
	var req map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}
	tutorialLinkInString, assertTutorialLinkPass := req["tutorialLink"].(string)
	if !assertTutorialLinkPass {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error, invaliable tutorialLink type")
		return
	}

	action, errInRetrieveAction := controller.Storage.ActionStorage.RetrieveByID(teamID, actionID)
	if errInRetrieveAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_ACTION, "get action failed: "+errInRetrieveAction.Error())
		return
	}

	action.SetTutorialLink(tutorialLinkInString, userID)

	// update
	errInUpdateAction := controller.Storage.ActionStorage.UpdateWholeAction(action)
	if errInUpdateAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "update action failed: "+errInUpdateAction.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
}
