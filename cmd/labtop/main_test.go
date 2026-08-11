package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunHelpDoesNotStartApplication(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	started := false

	code := run([]string{"--help"}, &stdout, &stderr, func() error {
		started = true
		return nil
	})

	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if started {
		t.Fatal("application started while displaying help")
	}
	if !strings.Contains(stdout.String(), "Usage: labtop") {
		t.Fatalf("help output = %q, want usage", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("error output = %q, want empty", stderr.String())
	}
}

func TestRunRejectsUnexpectedArgument(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"unexpected"}, &stdout, &stderr, func() error {
		t.Fatal("application started with an unexpected argument")
		return nil
	})

	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unexpected argument") {
		t.Fatalf("error output = %q, want unexpected argument", stderr.String())
	}
}

func TestRunReportsApplicationFailure(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, &stdout, &stderr, func() error {
		return errors.New("display failed")
	})

	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "display failed") {
		t.Fatalf("error output = %q, want application error", stderr.String())
	}
}
