package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/migsalazar/labtop/internal/config"
)

func TestRunHelpDoesNotLoadConfigurationOrStartApplication(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	loaded := false
	started := false
	code := run([]string{"--help"}, &stdout, &stderr,
		func(string) (config.Config, error) {
			loaded = true
			return config.Config{}, nil
		},
		func(config.Config) error {
			started = true
			return nil
		},
	)

	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if loaded || started {
		t.Fatalf("help loaded=%t started=%t, want both false", loaded, started)
	}
	if !strings.Contains(stdout.String(), "Usage: labtop [--config PATH]") {
		t.Fatalf("help output = %q, want usage", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("error output = %q, want empty", stderr.String())
	}
}

func TestRunPassesConfiguredPathAndConfiguration(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	want := config.Config{Console: config.Console{Title: "TEST"}}
	code := run([]string{"--config", "custom.toml"}, &stdout, &stderr,
		func(path string) (config.Config, error) {
			if path != "custom.toml" {
				t.Fatalf("config path = %q, want custom.toml", path)
			}
			return want, nil
		},
		func(got config.Config) error {
			if got.Console.Title != want.Console.Title {
				t.Fatalf("configuration title = %q, want %q", got.Console.Title, want.Console.Title)
			}
			return nil
		},
	)

	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("run() code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunRejectsExplicitEmptyConfigPath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--config="}, &stdout, &stderr,
		func(string) (config.Config, error) {
			t.Fatal("configuration loaded with an explicitly empty path")
			return config.Config{}, nil
		},
		func(config.Config) error {
			t.Fatal("application started with an explicitly empty path")
			return nil
		},
	)

	if code != 1 || !strings.Contains(stderr.String(), "path must not be empty") {
		t.Fatalf("run() code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunRejectsUnexpectedArgument(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"unexpected"}, &stdout, &stderr,
		func(string) (config.Config, error) {
			t.Fatal("configuration loaded with an unexpected argument")
			return config.Config{}, nil
		},
		func(config.Config) error {
			t.Fatal("application started with an unexpected argument")
			return nil
		},
	)

	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unexpected argument") {
		t.Fatalf("error output = %q, want unexpected argument", stderr.String())
	}
}

func TestRunReportsConfigurationFailureWithoutStartingApplication(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(nil, &stdout, &stderr,
		func(path string) (config.Config, error) {
			if path != "" {
				t.Fatalf("implicit path = %q, want empty", path)
			}
			return config.Config{}, errors.New("invalid layout")
		},
		func(config.Config) error {
			t.Fatal("application started with invalid configuration")
			return nil
		},
	)

	if code != 1 || !strings.Contains(stderr.String(), "configuration: invalid layout") {
		t.Fatalf("run() code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunReportsApplicationFailure(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(nil, &stdout, &stderr,
		func(string) (config.Config, error) { return config.Config{}, nil },
		func(config.Config) error { return errors.New("display failed") },
	)

	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "display failed") {
		t.Fatalf("error output = %q, want application error", stderr.String())
	}
}
