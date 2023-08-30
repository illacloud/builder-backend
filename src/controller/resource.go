package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/response"
	"github.com/illacloud/builder-backend/src/utils/accesscontrol"
	"github.com/illacloud/builder-backend/src/utils/auditlogger"
)

func (controller *Controller) GetAllResources(c *gin.Context) {
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
		accesscontrol.UNIT_TYPE_RESOURCE,
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
	resources, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamID(teamID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources by team id error: "+errInRetrieveResource.Error())
		return
	}

	// feedback
	c.JSON(http.StatusOK, model.BatchNewResourceForExport(resources))
	return
}

func (controller *Controller) CreateResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_CREATE_RESOURCE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	createResourceRequest := request.NewCreateResourceRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&createResourceRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate
	validate := validator.New()
	if err := validate.Struct(createResourceRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// new resource
	resource := model.NewResourceByCreateResourceRequest(teamID, userID, createResourceRequest)

	// validate options
	errInValidateResourceContent := controller.ValidateResourceConternt(c, resource)
	if errInValidateResourceContent != nil {
		return
	}

	// create
	_, errInCreateResource := controller.Storage.ResourceStorage.Create(resource)
	if errInCreateResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_RESOURCE, "create resources error: "+errInCreateResource.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:    auditlogger.AUDIT_LOG_CREATE_RESOURCE,
		TeamID:       teamID,
		UserID:       userID,
		IP:           c.ClientIP(),
		ResourceID:   resource.ID,
		ResourceName: resource.Name,
		ResourceType: resource.ExportTypeInString(),
	})

	// feedback
	controller.FeedbackOK(c, response.NewCreateResourceResponse(resource))
	return
}

func (controller *Controller) GetResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
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
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewGetResourceResponse(resource))
	return
}

func (controller *Controller) UpdateResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetUserID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
		accesscontrol.ACTION_MANAGE_EDIT_RESOURCE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// parse request body
	updateResourceRequest := request.NewUpdateResourceRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&updateResourceRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate update resource request
	validate := validator.New()
	if err := validate.Struct(updateResourceRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// get old resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}

	// update field
	resource.UpdateByUpdateResourceRequest(userID, updateResourceRequest)

	// validate options
	errInValidateResourceContent := controller.ValidateResourceConternt(c, resource)
	if errInValidateResourceContent != nil {
		return
	}

	// update database record
	errInUpdateResource := controller.Storage.ResourceStorage.UpdateWholeResource(resource)
	if errInUpdateResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_RESOURCE, "update resources error: "+errInUpdateResource.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:    auditlogger.AUDIT_LOG_UPDATE_RESOURCE,
		TeamID:       teamID,
		UserID:       userID,
		IP:           c.ClientIP(),
		ResourceID:   resource.ID,
		ResourceName: resource.Name,
		ResourceType: resource.ExportTypeInString(),
	})

	// feedback
	controller.FeedbackOK(c, response.NewUpdateResourceResponse(resource))
	return
}

func (controller *Controller) DeleteResource(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	canDelete, errInCheckAttr := controller.AttributeGroup.CanDelete(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		resourceID,
		accesscontrol.ACTION_DELETE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canDelete {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}

	// audit log
	auditLogger := auditlogger.GetInstance()
	auditLogger.Log(&auditlogger.LogInfo{
		EventType:    auditlogger.AUDIT_LOG_DELETE_RESOURCE,
		TeamID:       teamID,
		UserID:       userID,
		IP:           c.ClientIP(),
		ResourceID:   resourceID,
		ResourceName: resource.Name,
		ResourceType: resource.ExportTypeInString(),
	})

	// delete
	errInDeleteResource := controller.Storage.ResourceStorage.Delete(teamID, resourceID)
	if errInDeleteResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_DELETE_RESOURCE, "delete resources error: "+errInDeleteResource.Error())
		return
	}

	// feedback
	controller.FeedbackOK(c, response.NewDeleteResourceResponse(resourceID))
	return

}

func (controller *Controller) TestConnection(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	userID, errInGetUserID := controller.GetUserIDFromAuth(c)
	if errInGetTeamID != nil || errInGetAuthToken != nil || errInGetUserID != nil {
		return
	}

	// validate
	canManage, errInCheckAttr := controller.AttributeGroup.CanManage(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
		accesscontrol.DEFAULT_UNIT_ID,
		accesscontrol.ACTION_MANAGE_EDIT_RESOURCE,
	)
	if errInCheckAttr != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "error in check attribute: "+errInCheckAttr.Error())
		return
	}
	if !canManage {
		controller.FeedbackBadRequest(c, ERROR_FLAG_ACCESS_DENIED, "you can not access this attribute due to access control policy.")
		return
	}

	// format data to DTO struct
	testResourceConnectionRequest := request.NewTestResourceConnectionRequest()
	if err := json.NewDecoder(c.Request.Body).Decode(&testResourceConnectionRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_PARSE_REQUEST_BODY_FAILED, "parse request body error: "+err.Error())
		return
	}

	// validate request
	validate := validator.New()
	if err := validate.Struct(testResourceConnectionRequest); err != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate request body error: "+err.Error())
		return
	}

	// new temp resource
	resource := model.NewResourceByTestResourceConnectionRequest(teamID, userID, testResourceConnectionRequest)

	// test connection
	errInTestConnection := controller.TestResourceConnection(c, resource)
	if errInTestConnection != nil {
		return
	}

	// feedback
	controller.FeedbackOK(c, nil)
	return
}

func (controller *Controller) GetMetaInfo(c *gin.Context) {
	// fetch needed param
	teamID, errInGetTeamID := controller.GetMagicIntParamFromRequest(c, PARAM_TEAM_ID)
	resourceID, errInGetResourceID := controller.GetMagicIntParamFromRequest(c, PARAM_RESOURCE_ID)
	userAuthToken, errInGetAuthToken := controller.GetUserAuthTokenFromHeader(c)
	if errInGetTeamID != nil || errInGetResourceID != nil || errInGetAuthToken != nil {
		return
	}

	// validate
	canAccess, errInCheckAttr := controller.AttributeGroup.CanAccess(
		teamID,
		userAuthToken,
		accesscontrol.UNIT_TYPE_RESOURCE,
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

	// get resource
	resource, errInRetrieveResource := controller.Storage.ResourceStorage.RetrieveByTeamIDAndResourceID(teamID, resourceID)
	if errInRetrieveResource != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_RESOURCE, "get resources error: "+errInRetrieveResource.Error())
		return
	}

	// fetch meta info
	resourceMetaInfo, errInGetMetaInfo := controller.GetResourceMetaInfo(c, resource)
	if errInGetMetaInfo != nil {
		return
	}

	// feedback
	c.JSON(http.StatusOK, resourceMetaInfo)
	return
}
