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
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/sev"
)

func Test_loadEnv_PS_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://ssm.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		val := ""

		switch string(body) {
		case `{"Name":"/foo/bar/zoo","WithDecryption":true}`:
			val = "BAZ"
		case `{"Name":"/hoge/fuga/piyo","WithDecryption":true}`:
			val = "HOGERA"
		default:
			assert.Fail("unexpected secret id: " + string(body))
		}

		return httpmock.NewStringResponse(http.StatusOK, fmt.Sprintf(`{
			"Parameter": {
				"ARN":"arn:aws:ssm:us-east-1:123456789012:parameter<name>",
				"DataType":"text",
				"LastModifiedDate":0,
				"Name":"<name>",
				"Type":"SecureString",
				"Value":"%s",
				"Version":1
			}
		}`, val)), nil
	})

	envFrom := map[string]string{
		"FOO":   "parameterstore:///foo/bar/zoo",
		"PIYO":  "parameterstore:///hoge/fuga/piyo",
		"HELLO": "world",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSSMClient: func() (*ssm.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
			require.NoError(err)
			svc := ssm.NewFromConfig(cfg)
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

func Test_loadEnv_PS_Without_Prefix(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://ssm.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		val := ""

		switch string(body) {
		case `{"Name":"/foo/bar/zoo","WithDecryption":true}`:
			val = "BAZ"
		case `{"Name":"/hoge/fuga/piyo","WithDecryption":true}`:
			val = "HOGERA"
		default:
			assert.Fail("unexpected secret id: " + string(body))
		}

		return httpmock.NewStringResponse(http.StatusOK, fmt.Sprintf(`{
			"Parameter": {
				"ARN":"arn:aws:ssm:us-east-1:123456789012:parameter<name>",
				"DataType":"text",
				"LastModifiedDate":0,
				"Name":"<name>",
				"Type":"SecureString",
				"Value":"%s",
				"Version":1
			}
		}`, val)), nil
	})

	envFrom := map[string]string{
		"FOO":   "parameterstore://foo/bar/zoo",
		"PIYO":  "parameterstore://hoge/fuga/piyo",
		"HELLO": "world",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSSMClient: func() (*ssm.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
			require.NoError(err)
			svc := ssm.NewFromConfig(cfg)
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

func Test_loadEnv_PS_Err(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://ssm.us-east-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusServiceUnavailable, ""), nil
	})

	envFrom := map[string]string{
		"FOO":  "parameterstore:///foo/bar/zoo",
		"PIYO": "parameterstore:///hoge/fuga/piyo",
	}

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	providers := &mockProviders{
		newSSMClient: func() (*ssm.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc), config.WithRetryer(func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), 1)
			}))

			require.NoError(err)
			svc := ssm.NewFromConfig(cfg)
			return svc, nil
		},
	}

	_, err := sev.LoadEnv(envFrom, providers)
	assert.ErrorContains(err, "StatusCode: 503")
}
