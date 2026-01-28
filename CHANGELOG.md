# Changelog

All notable changes to Burrow will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-28

### Added

- Initial release of Burrow
- Deep system cleanup functionality
  - Temporary files cleanup
  - Browser cache removal (Chrome, Edge, Firefox, Brave)
  - Windows Update cache cleanup
  - Application cache cleanup
  - System logs cleanup
  - Recycle Bin cleanup
  - Thumbnail and icon cache cleanup
  - Prefetch cleanup
- Smart application uninstaller
  - Registry entry removal
  - AppData cleanup
  - Install location removal
  - Leftover detection
- Live system monitoring
  - CPU usage and per-core stats
  - Memory usage metrics
  - Disk usage and I/O stats
  - Network statistics
  - Top processes display
  - System health scoring
- System optimization features
  - DNS cache flush
  - Network adapter reset
  - Windows Search rebuild
  - Icon cache clear
  - Windows Update cleanup
  - SFC integrity check
  - Telemetry optimization
- Disk space analyzer
  - Recursive directory scanning
  - Visual tree display
  - Large file detection
  - Size-based filtering
- Interactive CLI menu
- Dry-run mode for safe previewing
- Debug mode for troubleshooting
- Whitelist management for protected paths
- Administrator privilege checks
- Windows-specific optimizations

### Security

- Safe file deletion with retry logic
- Registry access validation
- Path existence verification
- Administrator-only operations for system changes

### Documentation

- Comprehensive README
- Quick start guide
- Contributing guidelines
- MIT License

## [Unreleased]

### Planned Features

- GUI version (Electron wrapper)
- Scheduled cleanup tasks
- Cloud storage cleanup
- Duplicate file finder
- Drive health monitoring (S.M.A.R.T.)
- Export/import configuration
- Multi-language support
- PowerShell integration
- Windows Terminal integration
- Startup program manager UI
- Service optimization UI
- Registry cleaner with backup
- Restore point creation before cleanup
- Detailed cleanup reports
- Email notifications
- Integration with Windows Task Scheduler
