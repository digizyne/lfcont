package deployment

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Deploy() error {
	createCloudRunService := func(ctx *pulumi.Context) error {
		_, err := cloudrunv2.NewService(ctx, "automation-test-service-001", &cloudrunv2.ServiceArgs{
			Location: pulumi.String("us-central1"),
			Name:     pulumi.String("automation-test-service-001"),
			Scaling: &cloudrunv2.ServiceScalingArgs{
				MinInstanceCount: pulumi.Int(0),
				MaxInstanceCount: pulumi.Int(1),
				ScalingMode:      pulumi.String("AUTOMATIC"),
			},
			Template: &cloudrunv2.ServiceTemplateArgs{
				Containers: cloudrunv2.ServiceTemplateContainerArray{
					&cloudrunv2.ServiceTemplateContainerArgs{
						Image: pulumi.String("us-central1-docker.pkg.dev/jcastle-dev/local-first-public/test1:latest"),
					},
				},
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
		return fmt.Errorf("Failed to create or select stack: %v", err)
	}

	w := s.Workspace()
	err = w.InstallPlugin(ctx, "gcp", "v9.3.0")
	if err != nil {
		return fmt.Errorf("Failed to install program plugins: %v", err)
	}

	s.SetConfig(ctx, "gcp:project", auto.ConfigValue{Value: "jcastle-dev"})

	_, err = s.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("Failed to refresh stack: %v", err)
	}

	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	_, err = s.Up(ctx, stdoutStreamer)
	if err != nil {
		return fmt.Errorf("Failed to update stack: %v", err)
	}

	fmt.Println("Deployment succeeded!")
	return nil
}
