package main

import (
	"github.com/gin-gonic/gin"

	routes "github.com/digizyne/lfcont/internal/api"
)

func main() {
	router := gin.Default()
	routes.RegisterRoutes(router)
	router.Run("0.0.0.0:8080")
}
