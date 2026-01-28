# Burrow Quick Start Guide

## Installation (60 seconds)

### Method 1: Pre-built Binary (Easiest)

1. Download `wm.exe` from [Releases](https://github.com/zs0c131y/burrow/releases)
2. Move it to `C:\Windows\System32\` (or any folder in your PATH)
3. Open **PowerShell/CMD as Administrator**
4. Type `wm` and press Enter

### Method 2: Build from Source

```bash
git clone https://github.com/zs0c131y/burrow.git
cd burrow
build.bat    # On Windows
# OR
./build.sh   # On Linux/Mac (cross-compiles)
```

## First Run

Open PowerShell/CMD **as Administrator** and run:

```bash
wm
```

You'll see an interactive menu. Try these in order:

## 5-Minute Tutorial

### 1. Check System Health (30 sec)

```bash
wm status
```

See CPU, RAM, disk, network usage and health score. If health is below 70, time to optimize!

### 2. Preview Cleanup (1 min)

```bash
wm clean --dry-run
```

See what Burrow will clean **without deleting anything**. Typical finds:

- 5-20GB: Temp files
- 10-50GB: Browser caches
- 2-10GB: Windows Update cache

### 3. Deep Clean (2 min)

If preview looks good:

```bash
wm clean
```

Confirm with `y`. Watch gigabytes disappear!

### 4. Analyze Disk Space (1 min)

```bash
wm analyze
```

Find what's eating your storage. Try:

```bash
wm analyze C:\Users\YourName
```

### 5. Optimize System (30 sec)

```bash
wm optimize
```

Clears DNS cache, rebuilds search index, optimizes network settings.

## Daily Commands

```bash
# Quick health check
wm status

# Weekly cleanup
wm clean

# Find space hogs
wm analyze

# Remove unused apps
wm uninstall
```

## Pro Tips

### Always Preview First

```bash
wm clean --dry-run
```

### Protect Important Caches

```bash
wm clean --whitelist
```

### Monitor Continuously

```bash
wm status -w
```

Press Ctrl+C to stop.

### Clean Specific Categories

```bash
wm clean --categories browser,temp
```

Categories: `temp`, `cache`, `logs`, `browser`, `updates`

### Analyze Deeper

```bash
wm analyze C:\ -d 5 --min-size 100
```

Depth 5, show only files >100MB

## Troubleshooting

### "Access Denied"

**Solution**: Run as Administrator

- Right-click PowerShell/CMD
- Select "Run as Administrator"

### "wm is not recognized"

**Solution**: Add to PATH or use full path

```bash
C:\path\to\wm.exe status
```

### Not Finding Large Files

**Solution**: Run from Administrator terminal

```bash
wm analyze -d 4
```

## Safety Notes

✅ **Safe Operations:**

- `wm status` - Read-only, always safe
- `wm analyze` - Read-only, always safe
- `wm clean --dry-run` - Preview only
- `wm optimize --dry-run` - Preview only

⚠️ **Requires Admin:**

- `wm clean` - Deletes files
- `wm uninstall` - Removes apps
- `wm optimize` - Modifies system settings

## Next Steps

1. ⭐ Star the [GitHub repo](https://github.com/zs0c131y/burrow)
2. Read the [full README](README.md)
3. Join discussions for tips and tricks
4. Report bugs or request features

## Common Questions

**Q: Will this break Windows?**
A: No. Burrow only cleans temporary files and optional caches. Use `--dry-run` to preview.

**Q: How much space can I free?**
A: Typically 10-50GB on first run, 2-5GB on weekly maintenance.

**Q: Is it faster than CCleaner?**
A: Yes! Single executable, no bloat, CLI speed.

**Q: Can I schedule it?**
A: Yes! Use Windows Task Scheduler to run `wm clean` weekly.

**Q: Is my data safe?**
A: Burrow only touches temp files, caches, and system junk. Never touches documents or user files.

---

**Need help?** Open an issue on [GitHub](https://github.com/zs0c131y/burrow/issues)!
