package sev_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/sev"
)

type mockProviders struct {
	newSecretsManagerClient func() (*secretsmanager.Client, error)
	newSSMClient            func() (*ssm.Client, error)
}

func (p *mockProviders) NewSecretsManagerClient() (*secretsmanager.Client, error) {
	return p.newSecretsManagerClient()
}

func (p *mockProviders) NewSSMClient() (*ssm.Client, error) {
	return p.newSSMClient()
}

func Test_loadEnv_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://secretsmanager.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		val := ""

		switch string(body) {
		case `{"SecretId":"foo/bar/zoo"}`:
			val = "BAZ"
		case `{"SecretId":"hoge/fuga/piyo"}`:
			val = "HOGERA"
		default:
			assert.Fail("unexpected secret id: " + string(body))
		}

		return httpmock.NewStringResponse(http.StatusOK, fmt.Sprintf(`{
			"ARN":"arn:aws:secretsmanager:us-east-1:123456789012:secret:<secret-id>",
			"CreatedDate":0,
			"Name":"<secret-id>",
			"SecretString":"%s",
			"VersionId":"5048d25e-e46f-4a6c-87d9-b358e5c5dfcf",
			"VersionStages":["AWSCURRENT"]
		}`, val)), nil
	})

	envFrom := map[string]string{
		"FOO":   "secretsmanager://foo/bar/zoo",
		"PIYO":  "secretsmanager://hoge/fuga/piyo",
		"HELLO": "world",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSecretsManagerClient: func() (*secretsmanager.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
			require.NoError(err)
			svc := secretsmanager.NewFromConfig(cfg)
			return svc, nil
		},
	}

	value, err := sev.LoadEnv(envFrom, providers)
	require.NoError(err)
	assert.Equal(map[string]string{
		"FOO":   "BAZ",
		"PIYO":  "HOGERA",
		"HELLO": "world",
	}, value)
}

func Test_loadEnv_OK_JSON(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://secretsmanager.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		val := ""

		switch string(body) {
		case `{"SecretId":"foo/bar/zoo"}`:
			val = `{\"FOO\":\"BAR\",\"oof\":\"rab\"}`
		case `{"SecretId":"hoge/fuga/piyo"}`:
			val = `{\"FUGA\":\"FUGAFUGA\",\"gafu\":\"gafugafu\"}`
		default:
			assert.Fail("unexpected secret id: " + string(body))
		}

		return httpmock.NewStringResponse(http.StatusOK, fmt.Sprintf(`{
			"ARN":"arn:aws:secretsmanager:us-east-1:123456789012:secret:<secret-id>",
			"CreatedDate":0,
			"Name":"<secret-id>",
			"SecretString":"%s",
			"VersionId":"5048d25e-e46f-4a6c-87d9-b358e5c5dfcf",
			"VersionStages":["AWSCURRENT"]
		}`, val)), nil
	})

	envFrom := map[string]string{
		"FOO":   "secretsmanager://foo/bar/zoo:FOO",
		"PIYO":  "secretsmanager://hoge/fuga/piyo:FUGA",
		"HELLO": "world",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSecretsManagerClient: func() (*secretsmanager.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
			require.NoError(err)
			svc := secretsmanager.NewFromConfig(cfg)
			return svc, nil
		},
	}

	value, err := sev.LoadEnv(envFrom, providers)
	require.NoError(err)
	assert.Equal(map[string]string{
		"FOO":   "BAR",
		"PIYO":  "FUGAFUGA",
		"HELLO": "world",
	}, value)
}

func Test_loadEnv_Err(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://secretsmanager.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusServiceUnavailable, ""), nil
	})

	envFrom := map[string]string{
		"FOO":  "secretsmanager://foo/bar/zoo",
		"PIYO": "secretsmanager://hoge/fuga/piyo",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSecretsManagerClient: func() (*secretsmanager.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc), config.WithRetryer(func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), 1)
			}))

			require.NoError(err)
			svc := secretsmanager.NewFromConfig(cfg)
			return svc, nil
		},
	}

	_, err := sev.LoadEnv(envFrom, providers)
	assert.ErrorContains(err, "StatusCode: 503")
}

func Test_loadEnv_Err_NotFound(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://secretsmanager.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusNotFound, `{"__type":"ResourceNotFoundException","Message":"Secrets Manager can\\'t find the specified secret."}`), nil
	})

	envFrom := map[string]string{
		"FOO": "secretsmanager://foo/bar/zoo",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSecretsManagerClient: func() (*secretsmanager.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
			require.NoError(err)
			svc := secretsmanager.NewFromConfig(cfg)
			return svc, nil
		},
	}

	_, err := sev.LoadEnv(envFrom, providers)
	assert.ErrorContains(err, `ResourceNotFoundException: Secrets Manager can\'t find the specified secret`)
}

func Test_loadEnv_Without_AWS(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	envFrom := map[string]string{
		"FOO":   "BAR",
		"PIYO":  "HOGE",
		"HELLO": "world",
	}

	providers := &mockProviders{
		newSecretsManagerClient: func() (*secretsmanager.Client, error) {
			require.Fail("Must not call newSecretsManagerClient")
			return nil, nil
		},
	}

	value, err := sev.LoadEnv(envFrom, providers)
	require.NoError(err)
	assert.Equal(map[string]string{
		"FOO":   "BAR",
		"PIYO":  "HOGE",
		"HELLO": "world",
	}, value)
}
