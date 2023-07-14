package main

import (
	"os"

	"github.com/acorn-io/aws-acorn/pkg/config"
	"github.com/acorn-io/aws-acorn/pkg/context"
	"github.com/acorn-io/aws-acorn/pkg/service"
	"github.com/sirupsen/logrus"
)

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func main() {
	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	configFile := getEnv("CONFIG_FILE", "config.json")
	cdkOut := getEnv("CDK_CONTEXT_OUTFILE", "cdk.config.json")
	serviceOut := getEnv("SERVICE_OUTFILE", "service.json")

	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}

	contextData, err := context.Render(cfg)
	if err != nil {
		return err
	}

	serviceData, err := service.Render(cfg, contextData)
	if err != nil {
		return err
	}

	if err := config.WriteFile(cdkOut, contextData); err != nil {
		return err
	}

	return config.WriteFile(serviceOut, serviceData)
}
