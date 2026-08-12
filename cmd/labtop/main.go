package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/migsalazar/labtop/internal/app"
	"github.com/migsalazar/labtop/internal/config"
)

const description = "Display the small-display-first Labtop monitor."

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, config.Load, app.Run))
}

func run(
	args []string,
	stdout, stderr io.Writer,
	loadConfig func(string) (config.Config, error),
	runApp func(config.Config) error,
) int {
	flags := flag.NewFlagSet("labtop", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to a TOML configuration file")
	flags.Usage = func() {
		fmt.Fprintln(stdout, "Usage: labtop [--config PATH]")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, description)
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "labtop: unexpected argument: %s\n", flags.Arg(0))
		return 2
	}

	configFlagSet := false
	flags.Visit(func(current *flag.Flag) {
		if current.Name == "config" {
			configFlagSet = true
		}
	})
	if configFlagSet && *configPath == "" {
		fmt.Fprintln(stderr, "labtop: configuration: --config path must not be empty")
		return 1
	}

	configuration, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "labtop: configuration: %v\n", err)
		return 1
	}

	if err := runApp(configuration); err != nil {
		fmt.Fprintf(stderr, "labtop: %v\n", err)
		return 1
	}

	return 0
}
