package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/request"
	"github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

func (controller *Controller) ValidateActionTemplate(c *gin.Context, action *model.Action) error {
	if resourcelist.IsVirtualResourceHaveNoOption(action.ExportType()) {
		return nil
	}

	// check build
	actionFactory := model.NewActionFactoryByAction(action)
	actionAssemblyLine, errInBuild := actionFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return errInBuild
	}

	// check template
	_, errInValidate := actionAssemblyLine.ValidateActionTemplate(action.ExportTemplateInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action template error: "+errInValidate.Error())
		return errInValidate
	}
	return nil
}

func (controller *Controller) ValidateFlowActionTemplate(c *gin.Context, flowAction *model.FlowAction) error {
	if resourcelist.IsVirtualResourceHaveNoOption(flowAction.ExportType()) {
		return nil
	}

	// check build
	actionFactory := model.NewFlowActionFactoryByFlowAction(flowAction)
	actionAssemblyLine, errInBuild := actionFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return errInBuild
	}

	// check template
	_, errInValidate := actionAssemblyLine.ValidateActionTemplate(flowAction.ExportTemplateInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action template error: "+errInValidate.Error())
		return errInValidate
	}
	return nil
}

func (controller *Controller) createAction(c *gin.Context, teamID int, appID int, userID int, createActionRequest *request.CreateActionRequest) (*model.Action, error) {
	// append remote virtual resource (like aiagent, but the transformet is local virtual resource)
	if createActionRequest.IsRemoteVirtualAction() {
		// the AI_Agent need fetch resource info from resource manager, but illa drive does not need that
		if createActionRequest.NeedFetchResourceInfoFromSourceManager() {
			api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
			if errInNewAPI != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
				return nil, errInNewAPI
			}
			virtualResource, errInGetVirtualResource := api.GetResource(createActionRequest.ExportActionTypeInInt(), createActionRequest.ExportResourceIDInInt())
			if errInGetVirtualResource != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in fetch action mapped virtual resource: "+errInGetVirtualResource.Error())
				return nil, errInGetVirtualResource
			}
			createActionRequest.AppendVirtualResourceToTemplate(virtualResource)
		}
	}

	// get action mapped app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return nil, errInRetrieveApp
	}

	// init action instace
	action, errorInNewAction := model.NewAcitonByCreateActionRequest(app, userID, createActionRequest)
	if errorInNewAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "error in create action instance: "+errorInNewAction.Error())
		return nil, errorInNewAction
	}

	// validate action options
	errInValidateActionOptions := controller.ValidateActionTemplate(c, action)
	if errInValidateActionOptions != nil {
		return nil, errInValidateActionOptions
	}

	// create action
	_, errInCreateAction := controller.Storage.ActionStorage.Create(action)
	if errInCreateAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_CREATE_ACTION, "create action error: "+errInCreateAction.Error())
		return nil, errInCreateAction
	}

	// update app updatedAt, updatedBy, editedBy field
	app.Modify(userID)
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app modify info error: "+errInUpdateApp.Error())
		return nil, errInUpdateApp
	}

	return action, nil
}

func (controller *Controller) updateAction(c *gin.Context, teamID int, appID int, userID int, actionID int, updateActionRequest *request.UpdateActionRequest) (*model.Action, error) {
	// append remote virtual resource (like aiagent, but the transformet is local virtual resource)
	if updateActionRequest.IsRemoteVirtualAction() {
		// the AI_Agent need fetch resource info from resource manager, but illa drive does not need that
		if updateActionRequest.NeedFetchResourceInfoFromSourceManager() {
			api, errInNewAPI := illaresourcemanagersdk.NewIllaResourceManagerRestAPI()
			if errInNewAPI != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "error in fetch action mapped virtual resource: "+errInNewAPI.Error())
				return nil, errInNewAPI
			}
			virtualResource, errInGetVirtualResource := api.GetResource(updateActionRequest.ExportActionTypeInInt(), updateActionRequest.ExportResourceIDInInt())
			if errInGetVirtualResource != nil {
				controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "error in fetch action mapped virtual resource: "+errInGetVirtualResource.Error())
				return nil, errInGetVirtualResource
			}
			updateActionRequest.AppendVirtualResourceToTemplate(virtualResource)
		}
	}

	// get action mapped app
	app, errInRetrieveApp := controller.Storage.AppStorage.RetrieveAppByTeamIDAndAppID(teamID, appID)
	if errInRetrieveApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveApp.Error())
		return nil, errInRetrieveApp
	}

	// get action
	inDatabaseAction, errInRetrieveAction := controller.Storage.ActionStorage.RetrieveActionByTeamIDActionID(teamID, actionID)
	if errInRetrieveAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_GET_APP, "get app failed: "+errInRetrieveAction.Error())
		return nil, errInRetrieveAction
	}

	// update inDatabaseAction instance
	inDatabaseAction.UpdateAcitonByUpdateActionRequest(app, userID, updateActionRequest)

	// validate action options
	errInValidateActionOptions := controller.ValidateActionTemplate(c, inDatabaseAction)
	if errInValidateActionOptions != nil {
		return nil, errInValidateActionOptions
	}

	// update action
	errInUpdateAction := controller.Storage.ActionStorage.UpdateWholeAction(inDatabaseAction)
	if errInUpdateAction != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_ACTION, "update action error: "+errInUpdateAction.Error())
		return nil, errInUpdateAction
	}

	// update app updatedAt, updatedBy, editedBy field
	app.Modify(userID)
	errInUpdateApp := controller.Storage.AppStorage.UpdateWholeApp(app)
	if errInUpdateApp != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_UPDATE_APP, "update app modify info error: "+errInUpdateApp.Error())
		return nil, errInUpdateApp
	}

	// ok
	return inDatabaseAction, nil
}
