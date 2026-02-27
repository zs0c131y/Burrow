# Memory File

## Project: Burrow
- Windows-only CLI system optimizer written in Go 1.22+
- Single binary `wm.exe`, no databases, no external services
- Cross-compiles from Linux using `GOOS=windows GOARCH=amd64`

## Environment Setup
- Go 1.22.2 pre-installed at `/usr/bin/go`
- Dependencies downloaded via `go mod download`
- golangci-lint v1.61.0 installed at `~/bin/golangci-lint`
- Must use `GOOS=windows` for Windows cross-compile, vet, lint
- Platform-split files allow running tests natively on Linux

## Key Commands
- Build: `GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o wm.exe`
- Build all: `make build-all`
- Deps: `go mod download && go mod tidy`
- Format: `go fmt ./...`
- Vet: `GOOS=windows go vet ./...`
- Lint: `GOOS=windows ~/bin/golangci-lint run`
- Test (Linux): `go test -v ./pkg/... ./internal/analyzer/... ./internal/cleanup/...`
- Test (Windows): `GOOS=windows go test -v ./...`

## Production-Grade Changes Made
### Critical Bug Fixes
- FormatBytes: fixed index-out-of-bounds on values >= 1EB, clamped negative inputs
- CreateProgressBar: fixed panic from negative `strings.Repeat` counts
- cmd/status.go: fixed nil pointer dereference on hostInfo/memInfo when gopsutil fails
- cmd/analyze.go: fixed division by zero when totalSize == 0
- internal/uninstall: fixed removeRegistryEntries dropping first path segment (SOFTWARE)
- internal/uninstall: fixed registry root detection (was always LOCAL_MACHINE, now tracks source)
- internal/optimize: fixed resetNetwork leaving network broken on partial failure

### Completed Implementations
- Whitelist: full JSON config persistence in %APPDATA%/Burrow/whitelist.json
- ManageWhitelist: interactive add/remove/list
- GetOldestFiles: implemented with ModTime tracking
- GetDuplicates: replaced size-only matching with SHA-256 partial hash (8KB)

### Code Quality
- RequireAdmin now returns error instead of calling os.Exit (testable)
- Platform-specific code split using build tags
- 42 unit tests added across 4 packages
- All golangci-lint warnings eliminated (was 4, now 0)
- Proper error wrapping throughout
- Input validation on all CLI flags
