# Labtop

Labtop is an early Go terminal application intended for a small, always-on Raspberry Pi display. The current application is read-only and renders a truthful placeholder rather than simulated monitoring data.

## Current behavior

The current application:

- Opens in the terminal alternate screen
- Centers a placeholder when terminal dimensions are available
- Exits with `q` or `Ctrl+C`
- Provides `--help`
- Loads validated local TOML configuration before entering the alternate screen
- Supports a configured display title and a validated three-column arrangement of built-in module definitions
- Provides typed in-memory snapshot, status, event, bounded-history, formatting, and fixed-scale sparkline primitives for later monitoring stages
- Contains no collectors, network probes, live metrics, persistence, or external-process integration

Monitoring features are not implemented yet. The data primitives are not connected to the TUI, and configured modules are validated but are not rendered or collected yet.

## Configuration

Labtop accepts an optional configuration path:

```bash
go run ./cmd/labtop --config config.example.toml
```

Without `--config`, Labtop looks for `config.toml` in the current directory. If that file is absent, it uses generic built-in defaults. An explicitly requested file must exist and be valid.

Configuration is decoded strictly. Unknown fields, unsupported module types, mismatched settings blocks, invalid intervals, invalid machine ports, and overlapping or out-of-bounds grid positions fail before the TUI starts.

The built-in module registry currently recognizes `system`, `machines`, `agents`, and `events`. No collector or agent provider is implemented. External targets and machine definitions are configuration data only and are never contacted by the current application.

Use `config.example.toml` as the sanitized reference. Keep deployment-specific values in the ignored local `config.toml`. Do not put credentials, tokens, or secrets in configuration; Labtop does not accept or require them.

## Environment

- Minimum Go language version: 1.25
- Preferred project toolchain: Go 1.26.5
- Development platform verified: macOS ARM64
- Runtime platform verified: Raspberry Pi OS/Linux ARM64 on a Raspberry Pi 5
- Target display verified: 1024×600 terminal display
- Cross-build verified: Linux ARM64 with `CGO_ENABLED=0`

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

Labtop also runs directly without tmux. A tmux window shared by clients with different dimensions uses shared terminal geometry; separate sessions avoid display cropping or apparent alignment shifts.

## Local verification

```bash
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
```

`gofmt -l .` should produce no output. Tests require no deployment infrastructure or public internet access.
