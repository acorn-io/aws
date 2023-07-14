package service

import (
	"github.com/acorn-io/aws-acorn/pkg/config"
)

type Acornfile struct {
	Services map[string]Service `json:"services,omitempty"`
}

type Service struct {
	Default bool `json:"default,omitempty"`
	Data    Data `json:"data,omitempty"`
}

type Data struct {
	AccountID  string         `json:"accountID,omitempty"`
	VPCID      string         `json:"vpcID,omitempty"`
	Region     string         `json:"region,omitempty"`
	CDKContext map[string]any `json:"cdkContext,omitempty"`
}

func Render(cfg *config.Config, contextData map[string]any) (*Acornfile, error) {
	return &Acornfile{
		Services: map[string]Service{
			"context": {
				Default: true,
				Data: Data{
					AccountID:  cfg.AccountID,
					VPCID:      cfg.VPCID,
					Region:     cfg.Region,
					CDKContext: contextData,
				},
			},
		},
	}, nil
}
