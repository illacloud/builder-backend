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

	// teams routers
	teamsRouter.PATCH("/:teamID/apps/:appID", r.Controller.PublishAppToMarketplaceInternal)

	// ai agentrouters
	appRouter.POST("/fetchByIDs", r.Controller.GetAllAppListByIDInternal)
	appRouter.GET("/:appID", r.Controller.GetAppInternal)

}
