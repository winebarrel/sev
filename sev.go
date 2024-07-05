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
)

var (
	_stdin  io.Reader = os.Stdin
	_stdout io.Writer = os.Stdout
	_stderr io.Writer = os.Stderr
)

func Run(options *Options) error {
	envFrom, err := loadEnvFrom(options.Config, options.Profile)

	if err != nil {
		return err
	}

	env, err := loadEnv(envFrom)

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

func loadEnv(envFrom map[string]string) (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		return nil, err
	}

	env := map[string]string{}
	svc := secretsmanager.NewFromConfig(cfg)

	for name, from := range envFrom {
		vkey := ""

		if strings.Contains(from, ":") {
			idKey := strings.SplitN(from, ":", 2)
			from = idKey[0]
			vkey = idKey[1]
		}

		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(from),
		}

		output, err := svc.GetSecretValue(context.Background(), input)

		if err != nil {
			return nil, err
		}

		value := aws.ToString(output.SecretString)

		if vkey != "" {
			var jsonValue map[string]string
			err := json.Unmarshal([]byte(value), &jsonValue)

			if err != nil {
				return nil, fmt.Errorf("failed to parse '%s': %w", from, err)
			}

			vval, ok := jsonValue[vkey]

			if !ok {
				return nil, fmt.Errorf("key could not be found in '%s': '%s'", from, vkey)
			}

			value = vval
		}

		env[name] = value
	}

	return env, nil
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
