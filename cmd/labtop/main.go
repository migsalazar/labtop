package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/migsalazar/labtop/internal/app"
)

const description = "Display the small-display-first Labtop monitor."

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, app.Run))
}

func run(args []string, stdout, stderr io.Writer, runApp func() error) int {
	flags := flag.NewFlagSet("labtop", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stdout, "Usage: labtop")
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

	if err := runApp(); err != nil {
		fmt.Fprintf(stderr, "labtop: %v\n", err)
		return 1
	}

	return 0
}
