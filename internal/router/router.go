package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "messaging-server/docs"
	"messaging-server/internal/handler"
	"messaging-server/internal/scheduler"
)

// SetupRouter wires up the /start, /stop, /list endpoints.
func SetupRouter(cron *scheduler.Scheduler) *gin.Engine {
	r := gin.Default()

	r.GET("/", handler.Index)

	r.GET("/health", handler.Health)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/messaging", handler.SchedulerControl(cron))

			v1.GET("/list", handler.ListMessages)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
