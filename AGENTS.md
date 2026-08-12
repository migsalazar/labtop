# Project guidance

- Labtop targets Raspberry Pi OS/Linux and a 1024×600 terminal display; development may occur on macOS.
- The current implementation uses the Go version declared in `go.mod`, Bubble Tea, Lip Gloss, `go-toml/v2`, and `gopsutil/v4`.
- The current Linux ARM64 build is pure Go and works with `CGO_ENABLED=0`.
- Configuration is local TOML, decoded strictly and validated before the alternate terminal screen opens.
- Keep `config.example.toml` sanitized and generic. Keep real `config.toml`, `.env` files, credentials, tokens, secrets, and private infrastructure details out of Git.
- Shared monitoring values use typed immutable snapshots, normalized events, explicit optional values, deterministic formatting, and fixed-capacity in-memory history.
- The isolated system collector publishes typed local snapshots and threshold transitions with cancellation-aware scheduling, but the application does not start it yet.
- The current application has no active collectors, remote network probes, displayed live metrics, persistence, or external-process integration. Machine and external targets remain validation data only.
- Keep reusable source generic. Do not hardcode deployment-specific hostnames, addresses, machine roles, services, integrations, or identity.
- Collectors must use contexts, bounded scheduling, independent optional values, and cancellation-aware sends. Expected source failures must not terminate other collection.
- Bubble Tea `View` methods and rendering helpers must not perform network, filesystem, process, or other blocking work.
- Keep the console read-only. Do not add remote execution or administrative controls without an explicit design decision.
- Never display invented metrics, health, activity, or monitored state.
- Prefer small packages, concrete structs, narrow interfaces, table-driven tests, and straightforward control flow.
- Prefer the standard library when it is sufficient. Add dependencies only when the implementation requires them.
- Run `gofmt`, `go vet`, `go test`, the race detector when concurrency changes, and a production build after implementation changes.
- Keep `README.md` limited to current behavior and verified commands. Add focused tests for every implemented behavior.
