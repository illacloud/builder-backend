package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/model"
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
