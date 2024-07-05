package sev

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Run_OK(t *testing.T) {
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
			val = `{\"FUGA\":\"FUGAFUGA\"}`
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
		}'`, val)), nil
	})

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[profile1]
FOO = "foo/bar/zoo"
[profile2]
piyo = "hoge/fuga/piyo:FUGA"
`)
	tomlFile.Sync()

	defer func() {
		_stdout = os.Stdout
		_stderr = os.Stderr
	}()

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	{
		bufout := &bytes.Buffer{}
		buferr := &bytes.Buffer{}
		_stdout = bufout
		_stderr = buferr

		options := &Options{
			Config:  tomlFile.Name(),
			Profile: "profile1",
			Command: []string{"/bin/sh", "-c", "echo $FOO"},
		}

		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
		require.NoError(err)

		err = Run(cfg, options)
		require.NoError(err)

		assert.Equal("BAZ\n", bufout.String())
		assert.Empty(buferr.String())
	}

	{
		bufout := &bytes.Buffer{}
		buferr := &bytes.Buffer{}
		_stdout = bufout
		_stderr = buferr

		options := &Options{
			Config:  tomlFile.Name(),
			Profile: "profile2",
			Command: []string{"/bin/sh", "-c", "echo $piyo"},
		}

		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(hc))
		require.NoError(err)

		err = Run(cfg, options)
		require.NoError(err)

		assert.Equal("FUGAFUGA\n", bufout.String())
		assert.Empty(buferr.String())
	}
}
