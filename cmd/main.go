package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/digizyne/lfcont/internal/api"
	"github.com/digizyne/lfcont/internal/data"
	"github.com/digizyne/lfcont/tools"
)

func main() {
	pool, err := data.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	app := tools.InitializeApp(pool)
	router := gin.Default()
	api.RegisterRoutes(router, app)
	router.Run("0.0.0.0:8080")
}
