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

	logrus.Infof("Writing to %s:\n%s", file, data)
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to %s: %v", file, err)
	}
	return nil
}
