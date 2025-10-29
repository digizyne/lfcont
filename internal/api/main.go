package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Pool *pgxpool.Pool
}

func InitializeApp(router *gin.Engine, pool *pgxpool.Pool) {
	app := &App{Pool: pool}

	router.GET("/health", app.CheckHealth)

	apiv1 := router.Group("/api/v1")
	auth := apiv1.Group("/auth")
	auth.POST("/register", app.register)
	auth.POST("/login", app.login)
	apiv1.POST("/container-registry", app.pushToContainerRegistry)
	apiv1.POST("/deploy", app.deploy)
}
