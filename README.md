# Labtop

Labtop is an early Go terminal application intended for a small, always-on Raspberry Pi display. The current application is read-only and renders a truthful placeholder rather than simulated monitoring data.

## Current behavior

The current scaffold:

- Opens in the terminal alternate screen
- Centers a placeholder when terminal dimensions are available
- Exits with `q` or `Ctrl+C`
- Provides `--help`
- Contains no configuration, collectors, network probes, metrics, persistence, or external-process integration

Monitoring features are not implemented yet.

## Environment

- Minimum Go language version: 1.25
- Preferred project toolchain: Go 1.26.5
- Development platform currently verified: macOS ARM64
- Cross-build currently verified: Linux ARM64 with `CGO_ENABLED=0`
- Raspberry Pi runtime behavior has not yet been validated on the physical device

## Building and running

Begin in the project root with a supported Go toolchain available:

```bash
mkdir -p bin
go build -trimpath -o ./bin/labtop ./cmd/labtop
./bin/labtop
```

For development without creating a project binary:

```bash
go run ./cmd/labtop
```

Show command help:

```bash
go run ./cmd/labtop --help
```

## Linux ARM64 cross-build

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -o ./dist/labtop-linux-arm64 ./cmd/labtop
```

The resulting executable is statically linked and does not require a Go runtime on the target device.

## Running inside tmux

```bash
tmux new -s labtop
./bin/labtop
```

Detach with `Ctrl+B`, then `D`. Reattach with:

```bash
tmux attach -t labtop
```

Labtop also runs directly without tmux.

## Local verification

```bash
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
```

`gofmt -l .` should produce no output. Tests require no deployment infrastructure or public internet access.
