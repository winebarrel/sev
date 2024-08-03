package sev

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Providers struct {
	awsConfigOptFns      AWSConfigOptFns
	secretsmanagerClient *secretsmanager.Client
}

type ProviderssIface interface {
	NewSecretsManagerClient() (*secretsmanager.Client, error)
}

func NewProviders(fns AWSConfigOptFns) *Providers {
	return &Providers{
		awsConfigOptFns: fns,
	}
}

func (p *Providers) NewSecretsManagerClient() (*secretsmanager.Client, error) {
	if p.secretsmanagerClient == nil {
		cfg, err := config.LoadDefaultConfig(context.Background(), p.awsConfigOptFns...)

		if err != nil {
			return nil, err
		}

		p.secretsmanagerClient = secretsmanager.NewFromConfig(cfg)
	}

	return p.secretsmanagerClient, nil
}
