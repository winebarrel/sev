package sev_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/sev"
)

type mockGetParameterAPI func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)

func (m mockGetParameterAPI) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return m(ctx, params, optFns...)
}

func Test_getParameter_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	svc := mockGetParameterAPI(func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
		assert.Equal("/foo/bar/zoo", aws.ToString(params.Name))

		outout := &ssm.GetParameterOutput{
			Parameter: &types.Parameter{
				Value: aws.String("BAZ"),
			},
		}

		return outout, nil
	})

	value, err := sev.GetParameter(svc, "/foo/bar/zoo")
	require.NoError(err)
	assert.Equal("BAZ", value)
}

func Test_getParameter_Err(t *testing.T) {
	assert := assert.New(t)

	svc := mockGetParameterAPI(func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
		return nil, errors.New("unexpected error")
	})

	_, err := sev.GetParameter(svc, "/foo/bar/zoo")

	assert.ErrorContains(err, "unexpected error")
}
