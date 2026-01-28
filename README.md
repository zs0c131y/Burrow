# Burrow

*Dig deep like a mole to optimize your Windows system.*

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Windows-0078D6.svg)](https://www.microsoft.com/windows)

![Burrow](docs/banner.png)

## Features

Burrow is an **all-in-one Windows system optimizer** that combines the functionality of multiple tools into a single, lightweight CLI application:

1. **Deep System Cleanup** - Removes temp files, caches, browser data, Windows Update remnants, and junk to free up gigabytes
2. **Smart App Uninstaller** - Removes applications plus all their leftovers (registry entries, app data, shortcuts, services)
3. **Live System Monitor** - Real-time dashboard showing CPU, RAM, disk, network, and top processes with health scoring
4. **System Optimization** - Optimizes services, startup programs, network settings, and registry for better performance
5. **Disk Space Analyzer** - Visual tree explorer to identify space hogs and large files

## Quick Start

### Installation

**Prerequisites:**

- Windows 10/11
- Administrator privileges

#### Option 1: Download Binary (Recommended)

Download the latest release from [Releases](https://github.com/zs0c131y/burrow/releases) and add it to your PATH.

#### Option 2: Build from Source

```bash
git clone https://github.com/zs0c131y/burrow.git
cd burrow
go build -o wm.exe
```

#### Option 3: Install with Go

```bash
go install github.com/zs0c131y/burrow@latest
```

### Basic Usage

```bash
# Interactive menu (recommended for first-time users)
wm

# Individual commands
wm clean                    # Deep system cleanup
wm clean --dry-run          # Preview what will be cleaned
wm clean --whitelist        # Manage protected paths

wm uninstall                # Smart app removal with leftover cleanup

wm optimize                 # System performance optimization

wm status                   # Live system monitoring dashboard
wm status -w                # Continuous monitoring mode

wm analyze                  # Disk space analyzer
wm analyze C:\Users         # Analyze specific path
wm analyze -d 5             # Analyze with depth 5

wm version                  # Show version information
wm --help                   # Show help
```

### Running as Administrator

**Important:** Most Burrow commands require administrator privileges. Right-click your terminal and select "Run as Administrator" before running commands.

## Features in Detail

### 🧹 Deep System Cleanup

Comprehensive cleanup that targets:

**Temporary Files:**

- Windows Temp (`C:\Windows\Temp`)
- User Temp (`%TEMP%`, `%TMP%`)
- Local temporary storage

**Cache Files:**

- Application caches
- Icon cache
- Thumbnail cache
- Prefetch files

**Browser Data:**

- Chrome, Edge, Firefox, Brave caches
- Browser temporary files

**Windows Specific:**

- Windows Update download cache
- Delivery Optimization cache
- Windows Error Reporting
- System logs and event logs
- Recycle Bin

**Features:**

- Dry-run mode to preview changes
- Whitelist management for protected paths
- Category-specific cleanup
- Detailed progress reporting

```bash
# Preview cleanup
wm clean --dry-run

# Clean specific categories
wm clean --categories temp,cache,browser

# Manage protected paths
wm clean --whitelist
```

### 🗑️ Smart App Uninstaller

Unlike Windows built-in uninstaller, Burrow removes:

- Application executables and files
- Registry keys (both HKLM and HKCU)
- AppData and LocalAppData remnants
- ProgramData leftovers
- Start menu shortcuts
- Desktop shortcuts
- Scheduled tasks
- Windows services

```bash
wm uninstall

# Select from list of installed applications
# Automatically detects and removes all related files
```

### ⚡ System Optimization

Performance tuning operations:

**Network Optimization:**

- DNS cache flush
- Network adapter reset
- IP release/renew
- Winsock reset

**System Maintenance:**

- Windows Search index rebuild
- Icon cache clear
- System file integrity check (SFC)
- Component cleanup (DISM)

**Privacy & Performance:**

- Disable unnecessary telemetry
- Optimize Windows services
- Startup program management

```bash
wm optimize

# Dry-run mode available
wm optimize --dry-run
```

### 📊 Live System Status

Real-time monitoring dashboard showing:

**CPU Metrics:**

- Total CPU usage with visual bars
- Per-core breakdown
- CPU model and core count

**Memory Metrics:**

- Used/Free/Cached memory
- Memory usage percentage
- Available RAM

**Disk Metrics:**

- Disk usage per partition
- Free space available
- Read/Write I/O statistics

**Network Metrics:**

- Bytes sent/received
- Packet statistics
- Active connections

**Process Information:**

- Top 5 processes by CPU
- Memory usage per process
- Real-time updates

**System Health:**

- Health score (0-100)
- Color-coded alerts
- System uptime

```bash
# Single snapshot
wm status

# Continuous monitoring (updates every 2 seconds)
wm status -w

# Custom refresh interval
wm status -w -i 5
```

### 💾 Disk Space Analyzer

Visual disk usage explorer:

**Features:**

- Recursive directory scanning
- Size-based sorting
- Visual percentage bars
- Large file detection (>100MB)
- Configurable depth
- Hidden file support
- Minimum size filtering

**Display:**

- Top 20 space consumers
- Color-coded bars (green/yellow/red)
- File vs. folder icons
- Item counts for directories

```bash
# Analyze C: drive
wm analyze

# Analyze specific path
wm analyze "C:\Users\YourName\Documents"

# Deeper analysis (depth 5)
wm analyze -d 5

# Show hidden files
wm analyze --hidden

# Filter files smaller than 50MB
wm analyze --min-size 50
```

## Command Reference

### Global Flags

```bash
--debug         Enable debug mode with detailed logs
--dry-run       Preview changes without making them
--help          Show help for any command
```

### Clean Command

```bash
wm clean [flags]

Flags:
  --whitelist           Manage protected paths
  --categories strings  Specific categories (temp,cache,logs,browser,updates)
```

### Uninstall Command

```bash
wm uninstall

Interactive menu to select and remove applications
```

### Optimize Command

```bash
wm optimize [flags]

Optimizes system performance and settings
```

### Status Command

```bash
wm status [flags]

Flags:
  -w, --watch              Continuous monitoring mode
  -i, --interval int       Update interval in seconds (default 2)
```

### Analyze Command

```bash
wm analyze [path] [flags]

Flags:
  -d, --depth int          Maximum depth to analyze (default 3)
      --hidden             Show hidden files and folders
      --min-size int       Minimum size in MB to display
```

## Safety Features

1. **Administrator Check**: Prevents accidental runs without proper privileges
2. **Dry-Run Mode**: Preview all changes before executing
3. **Whitelist System**: Protect critical paths from cleanup
4. **Safe Delete**: Retry logic for locked files
5. **Error Handling**: Graceful handling of inaccessible paths
6. **Detailed Logging**: Debug mode for troubleshooting

## Performance

- **Binary Size**: ~8MB (single executable)
- **Memory Usage**: <50MB during operation
- **Cleanup Speed**: ~5-10GB per minute
- **Scan Speed**: ~100K files per minute

## Tips & Best Practices

1. **Always run as Administrator** for full functionality
2. **Use dry-run first** to preview changes: `wm clean --dry-run`
3. **Create restore point** before major optimizations
4. **Regular maintenance**: Run cleanup weekly for best results
5. **Monitor health**: Check `wm status` to identify performance issues
6. **Whitelist important caches**: Some application caches improve performance

## Comparison with Other Tools

| Feature             | Burrow | CCleaner | CleanMyPC | Windows Cleanup |
| ------------------- | ------ | -------- | --------- | --------------- |
| CLI Interface       | ✅     | ❌       | ❌        | ❌              |
| Free & Open Source  | ✅     | Freemium | ❌        | ✅              |
| Registry Cleanup    | ✅     | ✅       | ✅        | ❌              |
| Smart Uninstaller   | ✅     | Premium  | ✅        | ❌              |
| Live Monitoring     | ✅     | Premium  | ❌        | ❌              |
| Disk Analyzer       | ✅     | Premium  | ✅        | ✅              |
| System Optimization | ✅     | Premium  | ✅        | ❌              |
| Dry-Run Mode        | ✅     | ❌       | ❌        | ❌              |
| Single Binary       | ✅     | ❌       | ❌        | ✅              |

## Troubleshooting

### "Access Denied" Errors

**Solution**: Run terminal as Administrator

### Cleanup Not Freeing Space

**Solution**:

- Check if paths are whitelisted: `wm clean --whitelist`
- Run `wm clean --debug` to see detailed logs
- Some files may be in use - reboot and try again

### Application Not Found in Uninstaller

**Solution**: Some apps don't register in Windows uninstall registry. Use manual removal or portable app uninstaller.

### High Memory Usage During Analysis

**Solution**: Reduce depth: `wm analyze -d 2` or filter small files: `wm analyze --min-size 100`

## Building from Source

```bash
# Clone repository
git clone https://github.com/zs0c131y/burrow.git
cd burrow

# Install dependencies
go mod download

# Build
go build -o wm.exe

# Build with optimizations
go build -ldflags="-s -w" -o wm.exe

# Cross-compile (from Linux/Mac)
GOOS=windows GOARCH=amd64 go build -o wm.exe
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Roadmap

- [ ] GUI version (Electron wrapper)
- [ ] Scheduled cleanup tasks
- [ ] Cloud storage cleanup
- [ ] Duplicate file finder
- [ ] Drive health monitoring (S.M.A.R.T.)
- [ ] Export/import configuration
- [ ] Multi-language support
- [ ] PowerShell integration
- [ ] Windows Terminal integration

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

Inspired by [Mole](https://github.com/tw93/Mole) for macOS

Built with:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [gopsutil](https://github.com/shirou/gopsutil) - System monitoring
- [promptui](https://github.com/manifoldco/promptui) - Interactive prompts
- [color](https://github.com/fatih/color) - Terminal colors

## Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/zs0c131y/burrow/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/zs0c131y/burrow/discussions)
- ⭐ **Star the repo** if Burrow helped you!

## Connect

- **GitHub**: [@zs0c131y](https://github.com/zs0c131y)
- **Twitter/X**: [@brewcask](https://twitter.com/brewcask)
- **Website**: [adarshg.dev](https://adarshg.dev)

---

**Made with ❤️ by [Adarsh](https://adarshg.dev)**
