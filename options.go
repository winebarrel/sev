package sev

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
)

type Options struct {
	Config             string                            `required:"" default:"~/.sev.toml" env:"SEV_CONFIG" help:"Config file path."`
	Profile            string                            `arg:"" required:"" help:"Profile name."`
	Command            []string                          `arg:"" required:"" help:"Command and arguments."`
	OverrideAwsProfile bool                              `negatable:"" default:"true" help:"Use AWS_PROFILE in sev config. (enabled by default)"`
	AWSConfigOptFns    []func(*config.LoadOptions) error `kong:"-"`
}

func (options *Options) AfterApply() error {
	if strings.HasPrefix(options.Config, "~/") {
		home, err := os.UserHomeDir()

		if err == nil {
			options.Config = strings.Replace(options.Config, "~", home, 1)
		}
	}

	return nil
}
