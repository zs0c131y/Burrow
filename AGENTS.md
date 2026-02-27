# AGENTS.md

## Cursor Cloud specific instructions

### Project overview

Burrow is a Windows-only CLI system optimizer written in Go 1.22+. It cross-compiles from Linux to Windows. The binary is `wm.exe`. See `README.md` for full feature list and usage.

### Critical: Windows-only runtime

- All code uses `golang.org/x/sys/windows` and `windows/registry` imports. The binary **cannot run on Linux**.
- `cmd/root.go` checks `runtime.GOOS != "windows"` and exits immediately on non-Windows.
- Platform-specific code is split via build tags: `platform_windows.go` and `platform_other.go` in `pkg/utils/`.
- **Use `GOOS=windows` for vet and lint** (otherwise Windows-only imports fail).
- Tests can run natively on Linux (platform stubs provided).

### Build & dev commands

| Task | Command |
|------|---------|
| Download deps | `go mod download` |
| Cross-compile | `GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o wm.exe` |
| Build all targets | `make build-all` |
| Format | `go fmt ./...` |
| Vet | `GOOS=windows go vet ./...` |
| Lint | `GOOS=windows ~/bin/golangci-lint run` |
| Test (Linux) | `go test -v ./pkg/... ./internal/analyzer/... ./internal/cleanup/...` |
| Test (Windows) | `GOOS=windows go test -v ./...` |

See `Makefile` for additional targets.

### Gotchas

- `golangci-lint` is not pre-installed; install to `~/bin` with `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ~/bin`.
- The Makefile `lint` target does not set `GOOS=windows`; run lint manually with `GOOS=windows ~/bin/golangci-lint run`.
- `internal/uninstall` and `internal/optimize` contain Windows-only code (registry, exec of Windows commands) that cannot be tested on Linux.
- Whitelist config persists at `%APPDATA%/Burrow/whitelist.json`.
