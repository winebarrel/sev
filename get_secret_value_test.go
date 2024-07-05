package sev_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/sev"
)

type mockGetSecretValueAPI func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

func (m mockGetSecretValueAPI) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

func Test_getSecretValue_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	svc := mockGetSecretValueAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		assert.Equal("foo/bar/zoo", aws.ToString(params.SecretId))

		outout := &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("BAZ"),
		}

		return outout, nil
	})

	value, err := sev.GetSecretValue(svc, "foo/bar/zoo")

	require.NoError(err)
	assert.Equal("BAZ", value)
}

func Test_getSecretValue_OK_JSON(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	svc := mockGetSecretValueAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		assert.Equal("foo/bar/zoo", aws.ToString(params.SecretId))

		outout := &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(`{"HOGE":"FUGA","PIYO":"HOGERA"}`),
		}

		return outout, nil
	})

	{
		value, err := sev.GetSecretValue(svc, "foo/bar/zoo:HOGE")
		require.NoError(err)
		assert.Equal("FUGA", value)
	}

	{
		value, err := sev.GetSecretValue(svc, "foo/bar/zoo:PIYO")
		require.NoError(err)
		assert.Equal("HOGERA", value)
	}
}

func Test_getSecretValue_Err(t *testing.T) {
	assert := assert.New(t)

	svc := mockGetSecretValueAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		return nil, errors.New("unexpected error")
	})

	_, err := sev.GetSecretValue(svc, "foo/bar/zoo")

	assert.ErrorContains(err, "unexpected error")
}

func Test_getSecretValue_Err_JSON(t *testing.T) {
	assert := assert.New(t)

	svc := mockGetSecretValueAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		outout := &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(`{`),
		}

		return outout, nil
	})

	_, err := sev.GetSecretValue(svc, "foo/bar/zoo:HOGE")
	assert.ErrorContains(err, "failed to parse 'foo/bar/zoo'")
}

func Test_getSecretValue_Err_KeyNotFound(t *testing.T) {
	assert := assert.New(t)

	svc := mockGetSecretValueAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		assert.Equal("foo/bar/zoo", aws.ToString(params.SecretId))

		outout := &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(`{"HOGE":"FUGA","PIYO":"HOGERA"}`),
		}

		return outout, nil
	})

	_, err := sev.GetSecretValue(svc, "foo/bar/zoo:BAZ")
	assert.ErrorContains(err, "key could not be found in 'foo/bar/zoo': 'BAZ'")
}
