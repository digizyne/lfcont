package handlers

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"

	tools "github.com/digizyne/lfcont/tools"
)

func PushToContainerRegistry(c *gin.Context) {
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

	imageDetails, err := tools.GetContainerImageDetails(imageLoadResponse)
	if err != nil {
		log.Printf("Error getting image details: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get image details: %v", err),
		})
		return
	}

	imageID := imageDetails.ImageID
	imageName := imageDetails.ImageName

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
}
