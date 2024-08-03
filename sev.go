package sev

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	_stdin  io.Reader = os.Stdin
	_stdout io.Writer = os.Stdout
	_stderr io.Writer = os.Stderr
)

const (
	KeyAWSProfile        = "AWS_PROFILE"
	PrefixSecretsManager = "secretsmanager://"
	PrefixParameterStore = "parameterstore://"
)

type AWSConfigOptFns []func(*config.LoadOptions) error

func Run(options *Options) error {
	envFrom, err := loadEnvFrom(options.Config, options.Profile)

	if err != nil {
		return err
	}

	optFns := options.AWSConfigOptFns

	if options.OverrideAwsProfile {
		awsProfile, ok := envFrom[KeyAWSProfile]

		if ok {
			optFns = append(optFns, config.WithSharedConfigProfile(awsProfile))
		}
	}

	providers := NewProviders(optFns)
	env, err := loadEnv(envFrom, providers)

	if err != nil {
		return err
	}

	return execCmd(options.Command, env)
}

func loadEnvFrom(config string, profile string) (map[string]string, error) {
	var envFromByProfile map[string]map[string]string
	_, err := toml.DecodeFile(config, &envFromByProfile)

	if err != nil {
		return nil, err
	}

	envFrom, ok := envFromByProfile[profile]

	if !ok {
		return nil, fmt.Errorf("profile could not be found: %s", profile)
	}

	return envFrom, nil
}

func loadEnv(envFrom map[string]string, providers ProviderssIface) (map[string]string, error) {
	env := map[string]string{}

	for name, from := range envFrom {
		value := from

		if strings.HasPrefix(from, PrefixSecretsManager) {
			svc, err := providers.NewSecretsManagerClient()

			if err != nil {
				return nil, err
			}

			fromWitoutPrefix := strings.Replace(from, PrefixSecretsManager, "", 1)
			value, err = getSecretValue(svc, fromWitoutPrefix)

			if err != nil {
				return nil, fmt.Errorf("failed to get %s: %w", from, err)
			}
		} else if strings.HasPrefix(from, PrefixParameterStore) {
			svc, err := providers.NewSSMClient()

			if err != nil {
				return nil, err
			}

			fromWitoutPrefix := strings.Replace(from, PrefixParameterStore, "", 1)

			if !strings.HasPrefix(fromWitoutPrefix, "/") {
				fromWitoutPrefix = "/" + fromWitoutPrefix
			}

			value, err = getParameter(svc, fromWitoutPrefix)

			if err != nil {
				return nil, fmt.Errorf("failed to get %s: %w", from, err)
			}
		}

		env[name] = value
	}

	return env, nil
}

type SecretsManagerGetSecretValueAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func getSecretValue(api SecretsManagerGetSecretValueAPI, from string) (string, error) {
	vkey := ""

	if strings.Contains(from, ":") {
		idKey := strings.SplitN(from, ":", 2)
		from = idKey[0]
		vkey = idKey[1]
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(from),
	}

	output, err := api.GetSecretValue(context.Background(), input)

	if err != nil {
		return "", err
	}

	value := aws.ToString(output.SecretString)

	if vkey != "" {
		var jsonValue map[string]string
		err := json.Unmarshal([]byte(value), &jsonValue)

		if err != nil {
			return "", fmt.Errorf("failed to parse '%s': %w", from, err)
		}

		vval, ok := jsonValue[vkey]

		if !ok {
			return "", fmt.Errorf("key could not be found in '%s': '%s'", from, vkey)
		}

		value = vval
	}

	return value, nil
}

type SSMGetParameterAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

func getParameter(api SSMGetParameterAPI, from string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(from),
		WithDecryption: aws.Bool(true),
	}

	output, err := api.GetParameter(context.Background(), input)

	if err != nil {
		return "", err
	}

	value := aws.ToString(output.Parameter.Value)
	return value, nil
}

func execCmd(cmdArgs []string, extraEnv map[string]string) error {
	name := cmdArgs[0]
	args := []string{}

	if len(cmdArgs) >= 2 {
		args = cmdArgs[1:]
	}

	env := os.Environ()

	for name, value := range extraEnv {
		env = append(env, name+"="+value)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = _stdin
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr
	cmd.Env = env

	return cmd.Run()
}
