package api

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"

	"github.com/digizyne/lfcont/tools"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/google/go-containerregistry/pkg/authn"
)

func (app *App) pushToContainerRegistry(c *gin.Context) {
	ctx := context.Background()

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Docker client error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to Docker daemon. Is Docker running?",
		})
		return
	}
	defer cli.Close()

	// Read gzip stream from request body
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

	// Load image into Docker daemon
	imageLoadResponse, err := cli.ImageLoad(ctx, gzr, client.ImageLoadWithQuiet(true))
	if err != nil {
		log.Printf("Docker ImageLoad error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Docker ImageLoad failed. Is the tar archive a valid 'docker save' output? Error: %v", err),
		})
		return
	}
	defer imageLoadResponse.Body.Close()

	// Get image details (specifically image ID and name so that we can tag it)
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

	// Tag image for target registry
	arRepoUrl := os.Getenv("AR_REPO_URL")
	targetTag := fmt.Sprintf("%s/%s:latest", arRepoUrl, imageName)
	err = cli.ImageTag(ctx, imageID, targetTag)
	if err != nil {
		log.Printf("Docker ImageTag error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Docker ImageTag failed: %v", err),
		})
		return
	}

	// Get image from local Docker daemon
	imageRef, err := name.ParseReference(targetTag)
	if err != nil {
		log.Fatalf("Failed to parse source reference: %v", err)
	}
	img, err := daemon.Image(imageRef)
	if err != nil {
		log.Fatalf("Failed to read image from local Docker daemon. Ensure Docker is running and image '%s' exists. Error: %v", imageRef, err)
	}

	// Authenticate to Artifact Registry using Service Account key
	key, err := os.ReadFile("./sakey.json")
	if err != nil {
		log.Fatalf("Failed to read Service Account key file: %v", err)
	}
	auth := authn.FromConfig(authn.AuthConfig{
		Username: "_json_key",
		Password: string(key),
	})

	// Push image to Artifact Registry
	err = remote.Write(imageRef, img, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		log.Fatalf("Image push failed! Error: %v", err)
	}

	// Record pushed image in database
	_, err = app.Pool.Exec(ctx, `
			INSERT INTO container_images (fqin, username)
			VALUES ($1, $2)
		`, targetTag, "jfcastel00")
	if err != nil {
		log.Printf("DB insert error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to record image in database: %v", err),
		})
		return
	}

	log.Printf("Successfully pushed image %s to registry", targetTag)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Image successfully pushed to registry as %s", targetTag),
	})
}
