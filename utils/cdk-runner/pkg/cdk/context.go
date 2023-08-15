package cdk

import (
	"os"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/cdk/config"
	"github.com/acorn-io/aws/utils/cdk-runner/pkg/cdk/context"
)

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func GenerateCDKContext() error {
	cdkOut := getEnv("CDK_CONTEXT_OUTFILE", "cdk.config.json")

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		return err
	}

	contextData, err := context.Render(cfg)
	if err != nil {
		return err
	}

	return config.WriteFile(cdkOut, contextData)
}
