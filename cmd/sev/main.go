package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"
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
	err := sev.Run(options)

	if err != nil {
		log.Fatal(err)
	}
}
