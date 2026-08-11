# Project guidance

- Labtop targets Raspberry Pi OS/Linux and a 1024×600 terminal display; development may occur on macOS.
- The current implementation uses the Go version declared in `go.mod`, Bubble Tea, and Lip Gloss.
- The current Linux ARM64 build is pure Go and works with `CGO_ENABLED=0`.
- The current application is read-only and has no configuration, collectors, network probes, metrics, or external-process integration.
- Keep reusable source generic. Do not hardcode deployment-specific hostnames, addresses, machine roles, services, integrations, or identity.
- Keep credentials, tokens, secrets, and private infrastructure details outside reusable source.
- Bubble Tea `View` methods and rendering helpers must not perform network, filesystem, process, or other blocking work.
- Never display invented metrics, health, activity, or monitored state.
- Prefer small packages, concrete structs, narrow interfaces, table-driven tests, and straightforward control flow.
- Prefer the standard library when it is sufficient. Add dependencies only when the implementation requires them.
- Run `gofmt`, `go vet`, `go test`, the race detector when concurrency changes, and a production build after implementation changes.
- Keep `README.md` limited to current behavior and verified commands. Add focused tests for every implemented behavior.
