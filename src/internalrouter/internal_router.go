package internalrouter

import (
	"github.com/gin-gonic/gin"
	"github.com/illacloud/builder-backend/src/controller"
)

type Router struct {
	Controller *controller.Controller
}

func NewRouter(controller *controller.Controller) *Router {
	return &Router{
		Controller: controller,
	}
}

func (r *Router) RegisterRouters(engine *gin.Engine) {
	routerGroup := engine.Group("/api/v1")

	// init user group
	teamsRouter := routerGroup.Group("/teams")
	appRouter := routerGroup.Group("/apps")
	flowActionRouter := routerGroup.Group("/teams/:teamID/workflow/:workflowID/flowActions")
	flowActionRDuplicateouter := routerGroup.Group("/duplicateFlowActoins/fromTeamID/:fromTeamID/toTeamID/:toTeamID/fromWorkflowID/:fromWorkflowID/toWorkflowID/:toWorkflowID/fromVersion/:fromVersion/toVersion/:toVersion")

	// teams routers
	teamsRouter.PATCH("/:teamID/apps/:appID", r.Controller.PublishAppToMarketplaceInternal)

	// ai agentrouters
	appRouter.POST("/fetchByIDs", r.Controller.GetAllAppListByIDsInternal)
	appRouter.GET("/:appID/releaseVersion", r.Controller.GetReleaseVersionAppInternal)

	// flow action routers
	flowActionRouter.GET("/version/:version/all", r.Controller.GetWorkflowAllFlowActionsInternal)
	flowActionRouter.GET("/version/:version/type/:actionType", r.Controller.GetWorkflowFlowActionsByTypeInternal)
	flowActionRouter.GET("/id/:actionID", r.Controller.GetWorkflowFlowActionByIDInternal)
	flowActionRouter.POST("/:flowActionID/run", r.Controller.RunFlowActionInternal)
	flowActionRDuplicateouter.POST("", r.Controller.DuplicateFlowActionsInternal)

}
