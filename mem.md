# Memory File

## Project: Burrow
- Windows-only CLI system optimizer written in Go 1.22+
- Single binary `wm.exe`, no databases, no external services
- Cross-compiles from Linux using `GOOS=windows GOARCH=amd64`

## Environment Setup (completed)
- Go 1.22.2 pre-installed at `/usr/bin/go`
- Dependencies downloaded via `go mod download`
- golangci-lint v1.61.0 installed at `~/bin/golangci-lint`
- Must use `GOOS=windows` for all Go toolchain commands (build, vet, test, fmt, lint)
- No test files exist yet (`*_test.go`)

## Key Commands
- Build: `GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o wm.exe`
- Build all: `make build-all`
- Deps: `go mod download && go mod tidy`
- Format: `GOOS=windows go fmt ./...`
- Vet: `GOOS=windows go vet ./...`
- Lint: `GOOS=windows ~/bin/golangci-lint run`
- Test: `GOOS=windows go test -v ./...`

## Results
- Cross-compile: OK (PE32+ executable, 4.5M)
- go fmt: OK
- go vet: OK
- golangci-lint: 4 pre-existing warnings (errcheck)
- go test: OK (no test files, but compilation passes)
