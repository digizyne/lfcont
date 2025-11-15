package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"cloud.google.com/go/logging"
	"github.com/digizyne/lfcont/internal/api"
	"github.com/digizyne/lfcont/internal/data"
	"github.com/digizyne/lfcont/internal/middleware"
	"github.com/gin-contrib/cors"
)

func main() {
	pool, err := data.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	client, err := logging.NewClient(ctx, "local-first-476300")
	if err != nil {
		log.Fatalf("Could not initialize GCP logging: %v", err)
	}
	defer client.Close()
	gcpLogger := client.Logger("controller-logs-dev")

	corsConfig := cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	}

	router := gin.Default()
	router.Use(cors.New(corsConfig))
	router.Use(middleware.GcpLogger(gcpLogger))
	api.InitializeApp(router, pool)
	router.Run("0.0.0.0:8080")
}
