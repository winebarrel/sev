package sev

import (
	"os"
	"strings"
)

type Options struct {
	ConfigGlob         string          `required:"" default:"~/.sev.toml" env:"SEV_CONFIG" help:"Config file path glob pattern."`
	Profile            string          `arg:"" required:"" help:"Profile name."`
	Command            []string        `arg:"" required:"" help:"Command and arguments."`
	DefaultProfile     string          `env:"SEV_DEFAULT_PROFILE" help:"Fallback profile name."`
	OverrideAwsProfile bool            `negatable:"" default:"true" help:"Use AWS_PROFILE in sev config (enabled by default)."`
	AWSConfigOptFns    AWSConfigOptFns `kong:"-"`
}

func (options *Options) AfterApply() error {
	if strings.HasPrefix(options.ConfigGlob, "~/") {
		home, err := os.UserHomeDir()

		if err == nil {
			options.ConfigGlob = strings.Replace(options.ConfigGlob, "~", home, 1)
		}
	}

	return nil
}
