package cdk

import (
	"context"
	"os"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/aws/utils"
	"github.com/acorn-io/aws/utils/cdk-runner/pkg/cdk/config"
	cdkContext "github.com/acorn-io/aws/utils/cdk-runner/pkg/cdk/context"
)

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func GenerateCDKContext() error {
	cdkOut := getEnv("CDK_CONTEXT_OUTFILE", "cdk.context.json")

	ctx := context.Background()
	if err := utils.WaitForClientRole(ctx); err != nil {
		return err
	}

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		return err
	}

	contextData, err := cdkContext.Render(cfg)
	if err != nil {
		return err
	}

	return config.WriteFile(cdkOut, contextData)
}
