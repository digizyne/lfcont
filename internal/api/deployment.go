package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (app *App) deploy(c *gin.Context) {
	createCloudRunService := func(ctx *pulumi.Context) error {
		service, err := cloudrunv2.NewService(ctx, "automation-test-service-001", &cloudrunv2.ServiceArgs{
			Location: pulumi.String("us-central1"),
			Name:     pulumi.String("automation-test-service-001"),
			Template: &cloudrunv2.ServiceTemplateArgs{
				Scaling: &cloudrunv2.ServiceTemplateScalingArgs{
					MinInstanceCount: pulumi.Int(0),
					MaxInstanceCount: pulumi.Int(1),
				},
				Containers: cloudrunv2.ServiceTemplateContainerArray{
					&cloudrunv2.ServiceTemplateContainerArgs{
						Image: pulumi.String("us-central1-docker.pkg.dev/local-first-476300/container-images-dev/api:latest"),
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

		return nil
	}

	ctx := context.Background()

	s, err := auto.UpsertStackInlineSource(ctx, "dev", "testProject", createCloudRunService)
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

	// Log the resource changes for debugging
	log.Printf("Deployment completed successfully. Resource changes: %+v", resourceChanges)

	c.JSON(http.StatusOK, gin.H{
		"message":          "Deployment succeeded",
		"resource_changes": resourceChanges,
		"total_operations": totalChanges,
	})
}
