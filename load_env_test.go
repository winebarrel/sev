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
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/sev"
)

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
			val = fmt.Sprintf("unexpected secret id: %s", body)
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

	newSecretsManagerClient := func() (*secretsmanager.Client, error) {
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
		require.NoError(err)
		svc := secretsmanager.NewFromConfig(cfg)
		return svc, nil
	}

	value, err := sev.LoadEnv(envFrom, newSecretsManagerClient)
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
			val = fmt.Sprintf("unexpected secret id: %s", body)
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

	newSecretsManagerClient := func() (*secretsmanager.Client, error) {
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
		require.NoError(err)
		svc := secretsmanager.NewFromConfig(cfg)
		return svc, nil
	}

	value, err := sev.LoadEnv(envFrom, newSecretsManagerClient)
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

	newSecretsManagerClient := func() (*secretsmanager.Client, error) {
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc), config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 1)
		}))

		require.NoError(err)
		svc := secretsmanager.NewFromConfig(cfg)
		return svc, nil
	}

	_, err := sev.LoadEnv(envFrom, newSecretsManagerClient)
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

	newSecretsManagerClient := func() (*secretsmanager.Client, error) {
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
		require.NoError(err)
		svc := secretsmanager.NewFromConfig(cfg)
		return svc, nil
	}

	_, err := sev.LoadEnv(envFrom, newSecretsManagerClient)
	assert.ErrorContains(err, `ResourceNotFoundException: Secrets Manager can\'t find the specified secret`)
}
