# Burrow Installation Guide

Complete guide to installing and setting up Burrow on your Windows system.

## Table of Contents

- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
- [Building from Source](#building-from-source)
- [Adding to PATH](#adding-to-path)
- [Running as Administrator](#running-as-administrator)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Updating](#updating)
- [Uninstallation](#uninstallation)

## System Requirements

### Minimum Requirements

- Windows 10 (Build 1809) or later
- Windows 11 (all versions)
- 50MB free disk space
- Administrator privileges (for most features)

### Recommended

- Windows 10 21H2 or Windows 11
- 100MB free disk space
- PowerShell 5.1+ or PowerShell Core 7+

### Supported Architectures

- AMD64 (x86_64) - Primary
- x86 (32-bit) - Available

## Installation Methods

### Method 1: Download Pre-built Binary (Easiest)

**Step 1**: Download the latest release

Go to [Burrow Releases](https://github.com/zs0c131y/burrow/releases) and download `wm.exe`

**Step 2**: Move to a permanent location

```powershell
# Option A: System-wide (requires admin)
Move-Item wm.exe C:\Windows\System32\

# Option B: User-local
$userBin = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $userBin
Move-Item wm.exe $userBin\
```

**Step 3**: Verify installation

```powershell
wm version
```

### Method 2: Build from Source (Windows)

**Prerequisites:**

- Git for Windows
- Go 1.22 or later

**Step 1**: Install Go

Download from [golang.org/dl](https://golang.org/dl/) or use Chocolatey:

```powershell
choco install golang
```

**Step 2**: Clone repository

```powershell
git clone https://github.com/zs0c131y/burrow.git
cd burrow
```

**Step 3**: Build

```powershell
# Using build script
.\build.bat

# Or manually
go build -ldflags="-s -w" -o wm.exe
```

**Step 4**: Move to PATH (see Method 1, Step 2)

### Method 3: Install with Go

**Prerequisites:** Go 1.22+

```powershell
go install github.com/zs0c131y/burrow@latest
```

This installs to `%GOPATH%\bin` (usually `C:\Users\YourName\go\bin`)

Ensure this directory is in your PATH.

### Method 4: Package Managers (Future)

**Chocolatey** (Coming Soon)

```powershell
choco install burrow
```

**Scoop** (Coming Soon)

```powershell
scoop install burrow
```

**winget** (Coming Soon)

```powershell
winget install burrow
```

## Building from Source

### Windows

```powershell
# Clone
git clone https://github.com/zs0c131y/burrow.git
cd burrow

# Install dependencies
go mod download

# Build
.\build.bat

# Or with Make (if you have it)
make build
```

### Linux/macOS (Cross-compile)

```bash
# Clone
git clone https://github.com/zs0c131y/burrow.git
cd burrow

# Install dependencies
go mod download

# Cross-compile for Windows
./build.sh

# Or with Make
make build-linux
```

### Build Options

**Standard build:**

```bash
go build -o wm.exe
```

**Optimized build (smaller binary):**

```bash
go build -ldflags="-s -w" -o wm.exe
```

**With version info:**

```bash
go build -ldflags="-s -w -X main.version=1.0.0" -o wm.exe
```

## Adding to PATH

### Option 1: System-wide (Recommended)

Requires Administrator:

```powershell
# Copy to System32 (automatically in PATH)
Copy-Item wm.exe C:\Windows\System32\
```

### Option 2: User-specific

No admin required:

```powershell
# Create bin directory
$userBin = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $userBin

# Copy binary
Copy-Item wm.exe $userBin\

# Add to user PATH
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$userBin*") {
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$path;$userBin",
        "User"
    )
}

# Refresh PATH in current session
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

### Option 3: Project-specific

Keep in project directory and use full path:

```powershell
C:\path\to\burrow\wm.exe status
```

## Running as Administrator

### Method 1: Run terminal as admin

1. Search "PowerShell" or "CMD"
2. Right-click → "Run as Administrator"
3. Run `wm` commands normally

### Method 2: Elevate individual command

```powershell
Start-Process wm.exe -ArgumentList "clean" -Verb RunAs
```

### Method 3: Create admin shortcut

1. Right-click `wm.exe`
2. Create shortcut
3. Right-click shortcut → Properties
4. Advanced → "Run as administrator"

### When is Admin required?

**Requires Admin:**

- `wm clean` (deletes system files)
- `wm uninstall` (modifies registry)
- `wm optimize` (modifies services)

**No Admin needed:**

- `wm status` (read-only)
- `wm analyze` (read-only)
- `wm version` (read-only)
- Any command with `--dry-run`

## Verification

After installation, verify Burrow is working:

```powershell
# Check version
wm version

# Output should be:
# Burrow v1.0.0
# Windows System Optimizer
# Built with Go 1.22

# Run status (no admin needed)
wm status

# Test dry-run (no admin needed)
wm clean --dry-run
```

## Troubleshooting

### "wm is not recognized as an internal or external command"

**Solution**: Burrow is not in PATH

```powershell
# Check if file exists
Get-Command wm.exe -ErrorAction SilentlyContinue

# If not found, add to PATH (see above)
# Or use full path
C:\path\to\wm.exe version
```

### "Access is denied"

**Solution**: Need administrator privileges

1. Close current terminal
2. Open PowerShell/CMD as Administrator
3. Run command again

### "This app can't run on your PC"

**Solution**: Wrong architecture

- Download the correct version (AMD64 for 64-bit Windows, x86 for 32-bit)
- Check your Windows version: `systeminfo | findstr /B /C:"System Type"`

### DLL Load Failed

**Solution**: Missing Visual C++ Redistributable

Download and install:

- [Visual C++ Redistributable](https://aka.ms/vs/17/release/vc_redist.x64.exe)

### Go Build Errors

**Error**: `go: cannot find main module`

**Solution**: Run from project root where `go.mod` exists

**Error**: `undefined: windows.Token`

**Solution**: Update golang.org/x/sys package

```bash
go get -u golang.org/x/sys/windows
```

## Updating

### From Pre-built Binary

1. Download latest `wm.exe` from releases
2. Replace old binary
3. Verify: `wm version`

### From Source

```powershell
cd burrow
git pull origin main
.\build.bat
```

### Future: Auto-update

```powershell
wm update
```

## Uninstallation

### Complete Removal

```powershell
# Remove binary
Remove-Item C:\Windows\System32\wm.exe

# Or from user bin
Remove-Item $env:USERPROFILE\bin\wm.exe

# Remove from PATH if added manually
# (Edit environment variables)

# Optional: Remove config (future feature)
Remove-Item $env:APPDATA\burrow -Recurse
```

### Keep Configuration

Just remove the binary, configuration will be preserved.

## Advanced Setup

### PowerShell Profile Integration

Add to `$PROFILE`:

```powershell
# Burrow aliases
Set-Alias wms 'wm status'
Set-Alias wmc 'wm clean'
Set-Alias wma 'wm analyze'

# Auto-elevation function
function Invoke-BurrowAsAdmin {
    param($Command)
    Start-Process wm -ArgumentList $Command -Verb RunAs
}

# Usage: Invoke-BurrowAsAdmin "clean"
```

### Windows Terminal Integration

Add to `settings.json`:

```json
{
    "profiles": {
        "list": [
            {
                "name": "Burrow",
                "commandline": "powershell.exe -NoExit -Command wm",
                "icon": "path/to/icon.png",
                "elevate": true
            }
        ]
    }
}
```

### Task Scheduler (Automated Cleanup)

Create scheduled task:

```powershell
$action = New-ScheduledTaskAction -Execute "wm.exe" -Argument "clean"
$trigger = New-ScheduledTaskTrigger -Weekly -DaysOfWeek Sunday -At 3am
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
Register-ScheduledTask -TaskName "Burrow Weekly Cleanup" -Action $action -Trigger $trigger -Principal $principal
```

## Next Steps

After installation:

1. Read [QUICKSTART.md](QUICKSTART.md) for 5-minute tutorial
2. Run `wm` for interactive menu
3. Try `wm status` to check system health
4. Run `wm clean --dry-run` to preview cleanup

## Support

- 🐛 Report issues: [GitHub Issues](https://github.com/zs0c131y/burrow/issues)
- 💬 Ask questions: [GitHub Discussions](https://github.com/zs0c131y/burrow/discussions)
- 📖 Documentation: [README.md](README.md)

---

**Last Updated**: January 28, 2026
**Version**: 1.0.0
