package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"

	deployment "github.com/digizyne/lfcont/internal/deployment"
)

func main() {
	router := gin.Default()

	router.POST("container-registry", func(c *gin.Context) {
		ctx := context.Background()

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Printf("Docker client error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to connect to Docker daemon. Is Docker running?",
			})
			return
		}
		defer cli.Close()

		gzipStream := c.Request.Body

		if c.ContentType() != "application/gzip" {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be application/gzip"})
			return
		}

		gzr, err := gzip.NewReader(gzipStream)
		if err != nil {
			log.Printf("Gzip reader error: %v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Failed to create gzip reader (invalid gzip data).",
			})
			return
		}
		defer gzr.Close()

		imageLoadResponse, err := cli.ImageLoad(ctx, gzr, client.ImageLoadWithQuiet(true))
		if err != nil {
			log.Printf("Docker ImageLoad error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Docker ImageLoad failed. Is the tar archive a valid 'docker save' output? Error: %v", err),
			})
			return
		}
		defer imageLoadResponse.Body.Close()

		c.JSON(http.StatusOK, gin.H{
			"message": "Image pushed to container registry",
		})
	})

	router.POST("/deploy", func(c *gin.Context) {
		err := deployment.Deploy()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Deployment succeeded",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	router.Run()
}
