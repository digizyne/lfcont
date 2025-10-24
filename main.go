package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

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

		// Parse the ImageLoad response to get the loaded image ID
		var imageID string
		scanner := bufio.NewScanner(imageLoadResponse.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Loaded image ID:") {
				// Extract image ID from line like: {"stream":"Loaded image ID: sha256:abc123..."}
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(line), &result); err == nil {
					if stream, ok := result["stream"].(string); ok {
						if strings.HasPrefix(stream, "Loaded image ID: ") {
							imageID = strings.TrimPrefix(stream, "Loaded image ID: ")
							imageID = strings.TrimSpace(imageID)
							break
						}
					}
				}
			} else if strings.Contains(line, "Loaded image:") {
				// Alternative format: {"stream":"Loaded image: image_name:tag"}
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(line), &result); err == nil {
					if stream, ok := result["stream"].(string); ok {
						if strings.HasPrefix(stream, "Loaded image: ") {
							imageID = strings.TrimPrefix(stream, "Loaded image: ")
							imageID = strings.TrimSpace(imageID)
							break
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading ImageLoad response: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse image load response",
			})
			return
		}

		if imageID == "" {
			log.Printf("Could not extract image ID from ImageLoad response")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to determine loaded image ID",
			})
			return
		}

		// Tag the loaded image
		imageName := strings.Split(imageID, ":")[0]
		targetTag := fmt.Sprintf("us-central1-docker.pkg.dev/jcastle-dev/local-first-public/%s:latest", imageName)
		err = cli.ImageTag(ctx, imageID, targetTag)
		if err != nil {
			log.Printf("Docker ImageTag error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Docker ImageTag failed: %v", err),
			})
			return
		}

		err = exec.Command("docker", "push", targetTag).Run()
		if err != nil {
			log.Printf("Docker push error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Docker push failed: %v", err),
			})
			return
		}

		log.Printf("Successfully pushed image %s to registry", targetTag)
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Image successfully pushed to registry as %s", targetTag),
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
