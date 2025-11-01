package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/digizyne/lfcont/tools"
	"github.com/gin-gonic/gin"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type RequestBody struct {
	Name           string `json:"name"`
	ContainerImage string `json:"container_image"`
	Tier           string `json:"tier"`
}

func (app *App) deploy(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	userClaims, err := tools.GetUserClaims(authHeader)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: " + err.Error(),
		})
		return
	}
	log.Printf("Authenticated user: %s", userClaims.Username)

	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error":   "invalid request payload",
			"message": err.Error(),
		})
		return
	}

	createCloudRunService := func(ctx *pulumi.Context) error {
		service, err := cloudrunv2.NewService(ctx, req.Name, &cloudrunv2.ServiceArgs{
			Location:           pulumi.String("us-central1"),
			Name:               pulumi.String(req.Name),
			DeletionProtection: pulumi.Bool(false),
			Template: &cloudrunv2.ServiceTemplateArgs{
				Scaling: &cloudrunv2.ServiceTemplateScalingArgs{
					MinInstanceCount: pulumi.Int(0),
					MaxInstanceCount: pulumi.Int(1),
				},
				Containers: cloudrunv2.ServiceTemplateContainerArray{
					&cloudrunv2.ServiceTemplateContainerArgs{
						Image: pulumi.String(req.ContainerImage),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Allow public access using Cloud Run service IAM policy
		_, err = cloudrunv2.NewServiceIamBinding(ctx, "public-access", &cloudrunv2.ServiceIamBindingArgs{
			Project:  pulumi.String("local-first-476300"),
			Location: pulumi.String("us-central1"),
			Name:     service.Name,
			Role:     pulumi.String("roles/run.invoker"),
			Members: pulumi.StringArray{
				pulumi.String("allUsers"),
			},
		})
		if err != nil {
			return err
		}

		// Export the service URL as a stack output
		ctx.Export("serviceUrl", service.Uri)

		return nil
	}

	ctx := context.Background()

	// Create unique stack name to ensure each deployment creates a new service
	stackName := fmt.Sprintf("stack-%s-%s", userClaims.Username, req.Name)
	projectName := fmt.Sprintf("project-%s", req.Name)

	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, createCloudRunService)
	if err != nil {
		log.Printf("Stack creation error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create or select stack: %v", err),
		})
		return
	}

	w := s.Workspace()
	err = w.InstallPlugin(ctx, "gcp", "v9.3.0")
	if err != nil {
		log.Printf("Plugin install error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to install GCP plugin: %v", err),
		})
		return
	}

	s.SetConfig(ctx, "gcp:project", auto.ConfigValue{Value: "local-first-476300"})

	_, err = s.Refresh(ctx)
	if err != nil {
		log.Printf("Refresh error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to refresh stack: %v", err),
		})
		return
	}

	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	output, err := s.Up(ctx, stdoutStreamer)
	if err != nil {
		log.Printf("Deployment error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to update stack: %v", err),
		})
		return
	}

	// Check for errors in the deployment output
	if output.Summary.ResourceChanges == nil || len(*output.Summary.ResourceChanges) == 0 {
		log.Printf("No resource changes detected - possible deployment issue")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Deployment completed but no resources were changed",
		})
		return
	}

	// Check deployment result
	resourceChanges := *output.Summary.ResourceChanges
	totalChanges := 0
	for _, count := range resourceChanges {
		totalChanges += count
	}

	if totalChanges == 0 {
		log.Printf("No resource operations performed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Deployment completed but no resources were processed",
		})
		return
	}

	// Get the service URL from stack outputs
	outputs, err := s.Outputs(ctx)
	if err != nil {
		log.Printf("Failed to get stack outputs: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Deployment succeeded but failed to get service URL",
		})
		return
	}

	var serviceUrl string
	if urlOutput, exists := outputs["serviceUrl"]; exists {
		serviceUrl = urlOutput.Value.(string)
		log.Printf("Service deployed successfully at: %s", serviceUrl)
	} else {
		log.Printf("Warning: serviceUrl not found in stack outputs")
		serviceUrl = "URL not available"
	}

	// Log the resource changes for debugging
	log.Printf("Deployment completed successfully. Resource changes: %+v", resourceChanges)

	// Record new deployment in database
	_, err = app.Pool.Exec(ctx, `
			INSERT INTO deployments (name, url, tier, container_image, username)
			VALUES ($1, $2, $3, $4, $5)
		`, req.Name, serviceUrl, req.Tier, req.ContainerImage, userClaims.Username)
	if err != nil {
		log.Printf("DB insert error: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to record image in database: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_url": serviceUrl,
	})
}
