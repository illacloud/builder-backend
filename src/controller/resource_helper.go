package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

func (controller *Controller) ValidateResourceConternt(c *gin.Context, resource *model.Resource) error {
	if resourcelist.IsVirtualResourceHaveNoOption(resource.ExportType()) {
		return nil
	}

	// check build
	resourceFactory := model.NewActionFactoryByResource(resource)
	resourceAssemblyLine, errInBuild := resourceFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return errInBuild
	}

	// check template
	_, errInValidate := resourceAssemblyLine.ValidateResourceOptions(resource.ExportOptionsInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate resource option error: "+errInValidate.Error())
		return errInValidate
	}
	return nil
}

func (controller *Controller) TestResourceConnection(c *gin.Context, resource *model.Resource) error {
	if resourcelist.IsVirtualResourceHaveNoOption(resource.ExportType()) {
		return nil
	}

	// check build
	resourceFactory := model.NewActionFactoryByResource(resource)
	resourceAssemblyLine, errInBuild := resourceFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return errInBuild
	}

	// check template
	_, errInValidate := resourceAssemblyLine.ValidateResourceOptions(resource.ExportOptionsInMap())
	if errInValidate != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate resource option error: "+errInValidate.Error())
		return errInValidate
	}

	// test connection
	resourceConnection, errInTestConnection := resourceAssemblyLine.TestConnection(resource.ExportOptionsInMap())
	if errInTestConnection != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_TEST_RESOURCE_CONNECTION, "test resource connection error: "+errInTestConnection.Error())
		return errInTestConnection
	}
	if !resourceConnection.Success {
		errInConnection := errors.New("test resource connection error, resource connection failed")
		controller.FeedbackBadRequest(c, ERROR_FLAG_CAN_NOT_TEST_RESOURCE_CONNECTION, errInConnection.Error())
		return errInConnection
	}
	return nil
}

func (controller *Controller) GetResourceMetaInfo(c *gin.Context, resource *model.Resource) (*common.MetaInfoResult, error) {
	if resourcelist.IsVirtualResourceHaveNoOption(resource.ExportType()) {
		return nil, nil
	}

	// check build
	resourceFactory := model.NewActionFactoryByResource(resource)
	resourceAssemblyLine, errInBuild := resourceFactory.Build()
	if errInBuild != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "validate action type error: "+errInBuild.Error())
		return nil, errInBuild
	}

	// check template
	resourceMetaInfo, errInGetMetaInfo := resourceAssemblyLine.GetMetaInfo(resource.ExportOptionsInMap())
	if errInGetMetaInfo != nil {
		controller.FeedbackBadRequest(c, ERROR_FLAG_VALIDATE_REQUEST_BODY_FAILED, "get resource meta info error: "+errInGetMetaInfo.Error())
		return nil, errInGetMetaInfo
	}

	return &resourceMetaInfo, nil
}
