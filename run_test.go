package sev

import (
	"bytes"
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
		}`, val)), nil
	})

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[profile1]
FOO = "secretsmanager://foo/bar/zoo"
BAR = "baz"
[profile2]
piyo = "secretsmanager://hoge/fuga/piyo:FUGA"
HOGE = "PIYO"
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
			ConfigGlob: tomlFile.Name(),
			Profile:    "profile1",
			Command:    []string{"/bin/sh", "-c", "echo $FOO $BAR"},
		}

		options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
		err := Run(options)
		require.NoError(err)

		assert.Equal("BAZ baz\n", bufout.String())
		assert.Empty(buferr.String())
	}

	{
		bufout := &bytes.Buffer{}
		buferr := &bytes.Buffer{}
		_stdout = bufout
		_stderr = buferr

		options := &Options{
			ConfigGlob: tomlFile.Name(),
			Profile:    "profile2",
			Command:    []string{"/bin/sh", "-c", "echo $piyo $HOGE"},
		}

		options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
		err := Run(options)
		require.NoError(err)

		assert.Equal("FUGAFUGA PIYO\n", bufout.String())
		assert.Empty(buferr.String())
	}
}

func Test_Run_OK_Glob(t *testing.T) {
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
		}`, val)), nil
	})

	d := t.TempDir()
	os.WriteFile(d+"/foo.toml", []byte(`[profile1]
FOO = "secretsmanager://foo/bar/zoo"
BAR = "baz"
`), 0600)
	os.WriteFile(d+"/bar.toml", []byte(`[profile2]
piyo = "secretsmanager://hoge/fuga/piyo:FUGA"
HOGE = "PIYO"
`), 0600)

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
			ConfigGlob: d + "/*.toml",
			Profile:    "profile1",
			Command:    []string{"/bin/sh", "-c", "echo $FOO $BAR"},
		}

		options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
		err := Run(options)
		require.NoError(err)

		assert.Equal("BAZ baz\n", bufout.String())
		assert.Empty(buferr.String())
	}

	{
		bufout := &bytes.Buffer{}
		buferr := &bytes.Buffer{}
		_stdout = bufout
		_stderr = buferr

		options := &Options{
			ConfigGlob: d + "/*.toml",
			Profile:    "profile2",
			Command:    []string{"/bin/sh", "-c", "echo $piyo $HOGE"},
		}

		options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
		err := Run(options)
		require.NoError(err)

		assert.Equal("FUGAFUGA PIYO\n", bufout.String())
		assert.Empty(buferr.String())
	}
}

func Test_Run_OK_WithAWSProfile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://secretsmanager.us-west-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, `{
			"ARN":"arn:aws:secretsmanager:us-east-1:123456789012:secret:<secret-id>",
			"CreatedDate":0,
			"Name":"<secret-id>",
			"SecretString":"BAZ",
			"VersionId":"5048d25e-e46f-4a6c-87d9-b358e5c5dfcf",
			"VersionStages":["AWSCURRENT"]
		}`), nil
	})

	httpmock.RegisterResponder(http.MethodPost, "https://secretsmanager.ap-northeast-1.amazonaws.com/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, `{
			"ARN":"arn:aws:secretsmanager:us-east-1:123456789012:secret:<secret-id>",
			"CreatedDate":0,
			"Name":"<secret-id>",
			"SecretString":"{\"FUGA\":\"FUGAFUGA\"}",
			"VersionId":"5048d25e-e46f-4a6c-87d9-b358e5c5dfcf",
			"VersionStages":["AWSCURRENT"]
		}`), nil
	})

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[profile1]
FOO = "secretsmanager://foo/bar/zoo"
BAR = "baz"
AWS_PROFILE = "68011B7B-38B7-440D-8A86-4DFECC3ADD24"
[profile2]
piyo = "secretsmanager://hoge/fuga/piyo:FUGA"
HOGE = "PIYO"
AWS_PROFILE = "FE4395BA-5714-44B7-9FFF-822398A90FDB"
`)
	tomlFile.Sync()

	awsConfig, _ := os.CreateTemp("", "")
	defer os.Remove(awsConfig.Name())
	awsConfig.WriteString(`[profile 68011B7B-38B7-440D-8A86-4DFECC3ADD24]
region = "us-west-1"
[profile FE4395BA-5714-44B7-9FFF-822398A90FDB]
region = "ap-northeast-1"
`)
	awsConfig.Sync()

	awsSharedCredentials, _ := os.CreateTemp("", "")
	defer os.Remove(awsSharedCredentials.Name())
	awsSharedCredentials.WriteString(`[68011B7B-38B7-440D-8A86-4DFECC3ADD24]
aws_access_key_id = "dummy"
aws_secret_access_key = "dummy"
[FE4395BA-5714-44B7-9FFF-822398A90FDB]
aws_access_key_id = "dummy"
aws_secret_access_key = "dummy"
`)
	awsSharedCredentials.Sync()

	defer func() {
		_stdout = os.Stdout
		_stderr = os.Stderr
	}()

	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	t.Setenv("AWS_CONFIG_FILE", awsConfig.Name())
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", awsSharedCredentials.Name())

	{
		bufout := &bytes.Buffer{}
		buferr := &bytes.Buffer{}
		_stdout = bufout
		_stderr = buferr

		options := &Options{
			ConfigGlob:         tomlFile.Name(),
			Profile:            "profile1",
			Command:            []string{"/bin/sh", "-c", "echo $FOO $BAR"},
			OverrideAwsProfile: true,
		}

		options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
		err := Run(options)
		require.NoError(err)

		assert.Equal("BAZ baz\n", bufout.String())
		assert.Empty(buferr.String())
	}

	{
		bufout := &bytes.Buffer{}
		buferr := &bytes.Buffer{}
		_stdout = bufout
		_stderr = buferr

		options := &Options{
			ConfigGlob:         tomlFile.Name(),
			Profile:            "profile2",
			Command:            []string{"/bin/sh", "-c", "echo $piyo $HOGE"},
			OverrideAwsProfile: true,
		}

		options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
		err := Run(options)
		require.NoError(err)

		assert.Equal("FUGAFUGA PIYO\n", bufout.String())
		assert.Empty(buferr.String())
	}
}

func Test_Run_OK_WithoutAWSProfile(t *testing.T) {
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
		}`, val)), nil
	})

	tomlFile, _ := os.CreateTemp("", "")
	defer os.Remove(tomlFile.Name())
	tomlFile.WriteString(`[profile1]
FOO = "secretsmanager://foo/bar/zoo"
BAR = "baz"
[profile2]
piyo = "secretsmanager://hoge/fuga/piyo:FUGA"
HOGE = "PIYO"
`)
	tomlFile.Sync()

	defer func() {
		_stdout = os.Stdout
		_stderr = os.Stderr
	}()

	awsConfig, _ := os.CreateTemp("", "")
	defer os.Remove(awsConfig.Name())
	awsConfig.WriteString(`[profile 68011B7B-38B7-440D-8A86-4DFECC3ADD24]
region = "us-west-1"
[profile FE4395BA-5714-44B7-9FFF-822398A90FDB]
region = "ap-northeast-1"
`)
	awsConfig.Sync()

	awsSharedCredentials, _ := os.CreateTemp("", "")
	defer os.Remove(awsSharedCredentials.Name())
	awsSharedCredentials.WriteString(`[68011B7B-38B7-440D-8A86-4DFECC3ADD24]
aws_access_key_id = "dummy"
aws_secret_access_key = "dummy"
[FE4395BA-5714-44B7-9FFF-822398A90FDB]
aws_access_key_id = "dummy"
aws_secret_access_key = "dummy"
`)
	awsSharedCredentials.Sync()

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	t.Setenv("AWS_CONFIG_FILE", awsConfig.Name())
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", awsSharedCredentials.Name())

	bufout := &bytes.Buffer{}
	buferr := &bytes.Buffer{}
	_stdout = bufout
	_stderr = buferr

	options := &Options{
		ConfigGlob:         tomlFile.Name(),
		Profile:            "profile1",
		Command:            []string{"/bin/sh", "-c", "echo $FOO $BAR"},
		OverrideAwsProfile: true,
	}

	options.AWSConfigOptFns = append(options.AWSConfigOptFns, config.WithHTTPClient(hc))
	err := Run(options)
	require.NoError(err)

	assert.Equal("BAZ baz\n", bufout.String())
	assert.Empty(buferr.String())
}
