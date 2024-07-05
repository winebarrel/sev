package main

import (
	"context"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/winebarrel/sev"
)

var version string

func init() {
	log.SetFlags(0)
}

func parseArgs() *sev.Options {
	var CLI struct {
		sev.Options
		Version kong.VersionFlag
	}

	parser := kong.Must(&CLI, kong.Vars{"version": version})
	parser.Model.HelpFlag.Help = "Show help."
	_, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	return &CLI.Options
}

func main() {
	options := parseArgs()
	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	err = sev.Run(cfg, options)

	if err != nil {
		log.Fatal(err)
	}
}
