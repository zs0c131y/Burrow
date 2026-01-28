# Contributing to Burrow

Thank you for considering contributing to Burrow! This guide will help you get started.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## How to Contribute

### Reporting Bugs

1. Check if the bug already exists in [Issues](https://github.com/zs0c131y/burrow/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Burrow version (`wm version`)
   - Windows version
   - Debug output (`wm [command] --debug`)

### Suggesting Features

1. Check [Issues](https://github.com/zs0c131y/burrow/issues) for existing suggestions
2. Create a new issue with:
   - Clear use case
   - Expected behavior
   - Potential implementation ideas (optional)

### Pull Requests

#### Setup Development Environment

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/burrow.git
cd burrow

# Add upstream remote
git remote add upstream https://github.com/zs0c131y/burrow.git

# Install dependencies
go mod download
```

#### Making Changes

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Make your changes
# ... edit files ...

# Format code
go fmt ./...

# Test your changes
go test ./...

# Build to ensure it compiles
go build -o wm.exe
```

#### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write clear, descriptive comments
- Keep functions focused and small
- Use meaningful variable names

**Example:**

```go
// Good
func calculateDiskUsage(path string) (int64, error) {
    // Implementation
}

// Avoid
func calc(p string) (int64, error) {
    // Implementation
}
```

#### Commit Messages

Use clear, descriptive commit messages:

```text
feat: Add registry cleanup to optimize command
fix: Handle access denied errors in cleanup
docs: Update README with new analyze flags
refactor: Extract cleanup logic into separate package
test: Add tests for disk analyzer
```

Prefixes:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

#### Submitting Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
# Include:
# - Description of changes
# - Related issue number
# - Testing performed
```

### Testing

**Manual Testing:**

1. Build the binary: `go build -o wm.exe`
2. Run as Administrator
3. Test your changes:

   ```bash
   wm clean --dry-run
   wm status
   # etc.
   ```

**Automated Testing:**

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/cleanup
```

**Testing Checklist:**

- [ ] Commands run without errors
- [ ] Admin checks work correctly
- [ ] Dry-run mode works
- [ ] Debug output is helpful
- [ ] Edge cases handled
- [ ] No memory leaks

## Project Structure

```text
burrow/
├── cmd/                    # CLI commands
│   ├── root.go            # Main command & menu
│   ├── clean.go           # Cleanup command
│   ├── uninstall.go       # Uninstall command
│   ├── optimize.go        # Optimize command
│   ├── status.go          # Status monitoring
│   └── analyze.go         # Disk analyzer
├── internal/              # Internal packages
│   ├── cleanup/           # Cleanup logic
│   ├── uninstall/         # Uninstall logic
│   ├── optimize/          # Optimization logic
│   ├── monitor/           # System monitoring
│   └── analyzer/          # Disk analysis
├── pkg/                   # Public packages
│   ├── models/            # Data models
│   └── utils/             # Utilities
├── main.go               # Entry point
├── go.mod                # Dependencies
└── README.md             # Documentation
```

## Development Guidelines

### Adding a New Command

1. Create `cmd/yourcommand.go`
2. Implement the command using Cobra
3. Add to `cmd/root.go` menu
4. Create internal package if complex logic needed
5. Update README and QUICKSTART
6. Add tests

**Template:**

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/fatih/color"
)

var yourCmd = &cobra.Command{
    Use:   "yourcommand",
    Short: "Brief description",
    Long:  "Detailed description",
    Run: func(cmd *cobra.Command, args []string) {
        runYourCommand()
    },
}

func init() {
    rootCmd.AddCommand(yourCmd)
}

func runYourCommand() {
    // Implementation
}
```

### Adding a Cleanup Target

Edit `internal/cleanup/manager.go`:

```go
func (cm *CleanupManager) getYourTargets(paths map[string]string) []*models.CleanupTarget {
    return []*models.CleanupTarget{
        {
            Name:        "Your Target",
            Path:        filepath.Join(paths["LOCALAPPDATA"], "YourPath"),
            Description: "Description",
            Category:    models.CategoryYourCategory,
        },
    }
}

// Add to DiscoverTargets()
targets = append(targets, cm.getYourTargets(sysPaths)...)
```

### Adding System Monitoring Metrics

Edit `cmd/status.go` and add your metric display function.

## Release Process

1. Update version in `main.go`
2. Update CHANGELOG.md
3. Create git tag: `git tag v1.x.x`
4. Push tag: `git push --tags`
5. GitHub Actions will build releases

## Questions?

- Open a discussion on GitHub
- Comment on relevant issues
- Contact maintainers

## Recognition

Contributors will be:

- Listed in CONTRIBUTORS.md
- Credited in release notes
- Acknowledged in README

Thank you for contributing! 🎉
