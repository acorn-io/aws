package context

import (
	"github.com/acorn-io/aws-acorn/pkg/config"
	"github.com/sirupsen/logrus"
)

func Render(cfg *config.Config) (map[string]any, error) {
	cdkContext, err := NewContext(cfg.AccountID, cfg.Region)
	if err != nil {
		logrus.Fatal(err)
	}

	vpcPlugin := NewVpcPlugin(cfg.VPCID)
	cdkContext.AddPlugin(vpcPlugin)

	return ToData(cdkContext)
}

func ToData(cdkData *CdkContext) (map[string]any, error) {
	data := map[string]any{}

	for _, plugin := range cdkData.Plugins {
		content, err := plugin.Render(cdkData)
		if err != nil {
			return nil, err
		}
		for k, v := range content {
			data[k] = v
		}
	}

	return data, nil
}
