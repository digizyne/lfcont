package api

import (
	"github.com/gin-gonic/gin"

	"github.com/digizyne/lfcont/internal/api/handlers"
	"github.com/digizyne/lfcont/tools"
)

func RegisterRoutes(router *gin.Engine, appRouter *tools.AppRouter) {
	router.GET("/ping", handlers.CheckHealth(appRouter))

	// apiv1 := router.Group("/api/v1")
	// apiv1.POST("/container-registry", handlers.PushToContainerRegistry)
	// apiv1.POST("/deploy", handlers.Deploy)
}
