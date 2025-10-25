package main

import (
	"log"

	"github.com/gin-gonic/gin"

	routes "github.com/digizyne/lfcont/internal/api"
	data "github.com/digizyne/lfcont/internal/data"
)

func main() {
	pool, err := data.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()
	router := gin.Default()
	routes.RegisterRoutes(router)
	router.Run("0.0.0.0:8080")
}
