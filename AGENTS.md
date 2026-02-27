# AGENTS.md

## Cursor Cloud specific instructions

### Project overview

Burrow is a Windows-only CLI system optimizer written in Go 1.22+. It cross-compiles from Linux to Windows. The binary is `wm.exe`. See `README.md` for full feature list and usage.

### Critical: Windows-only runtime

- All code uses `golang.org/x/sys/windows` and `windows/registry` imports. The binary **cannot run on Linux**.
- `cmd/root.go` checks `runtime.GOOS != "windows"` and exits immediately on non-Windows.
- **You must use `GOOS=windows` for all Go commands** (`go build`, `go vet`, `go test`, `go fmt`, `golangci-lint run`), otherwise compilation fails due to Windows-only imports in `pkg/utils/helpers.go`.

### Build & dev commands

| Task | Command |
|------|---------|
| Download deps | `go mod download` |
| Cross-compile | `GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o wm.exe` |
| Build all targets | `make build-all` |
| Format | `GOOS=windows go fmt ./...` |
| Vet | `GOOS=windows go vet ./...` |
| Lint | `GOOS=windows golangci-lint run` (install: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \| sh -s -- -b ~/bin`) |
| Test | `GOOS=windows go test -v ./...` |

See `Makefile` for additional targets (`make deps`, `make build-linux`, `make fmt`, `make lint`, `make test`).

### Gotchas

- `golangci-lint` is not pre-installed; the Makefile `lint` target assumes it's on PATH. Install to `~/bin` and add to PATH.
- No `*_test.go` files exist yet. `go test ./...` compiles all packages but finds no tests.
- The Makefile `lint` target does not set `GOOS=windows`; run lint manually with `GOOS=windows golangci-lint run`.
- Pre-existing lint warnings (4 unchecked error returns) are in the codebase; these are not blockers.
- `build.sh` is the Unix cross-compile convenience script; `build.bat` is for building on Windows.
