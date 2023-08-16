package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Config struct {
	AccountID string `json:"accountID,omitempty"`
	Region    string `json:"region,omitempty"`
	VPCID     string `json:"vpcID,omitempty"`
}

func NewConfigFromEnv() (*Config, error) {
	config := &Config{}
	config.AccountID = os.Getenv("CDK_DEFAULT_ACCOUNT")
	config.Region = os.Getenv("CDK_DEFAULT_REGION")
	config.VPCID = os.Getenv("VPC_ID")

	return config, nil
}

func ReadConfig(config string) (*Config, error) {
	result := Config{}
	f, err := os.Open(config)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %v", config, err)
	}
	defer f.Close()

	return &result, json.NewDecoder(f).Decode(&result)
}

func WriteFile(file string, obj any) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	logrus.Infof("Writing cdk context to %s", file)
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		logrus.Errorf("error writing to %s: %v\ncontent\n%s", file, err, string(data))
		return err
	}
	return nil
}
