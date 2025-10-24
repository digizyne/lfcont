package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	handlers "github.com/digizyne/lfcont/internal/api/handlers"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ping": "pong",
		})
	})

	apiv1 := router.Group("/api/v1")
	apiv1.POST("/container-registry", handlers.PushToContainerRegistry)
	apiv1.POST("/deploy", handlers.Deploy)
}
