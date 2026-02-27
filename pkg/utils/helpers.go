package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FormatBytes converts bytes to human-readable format.
// Handles zero, negative (which shouldn't occur, but guards against int64 overflow),
// and extremely large values safely.
func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	div := int64(unit)
	exp := 0

	for n := bytes / unit; n >= unit && exp < len(units)-1; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatDuration converts duration to human-readable format.
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	hours := d.Hours()
	if hours < 24 {
		return fmt.Sprintf("%.1fh", hours)
	}
	days := int(hours / 24)
	remainingHours := int(hours) % 24
	return fmt.Sprintf("%dd %dh", days, remainingHours)
}

// GetDirSize calculates total size of a directory recursively.
// Returns (totalBytes, fileCount, error). Inaccessible files are silently skipped.
func GetDirSize(path string) (int64, int, error) {
	var size int64
	var count int

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	return size, count, err
}

// PathExists checks if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExpandEnvPath expands environment variables in a path.
func ExpandEnvPath(path string) string {
	if strings.Contains(path, "%") {
		return os.ExpandEnv(path)
	}
	return path
}

// RemoveDirectory safely removes a directory and its contents.
func RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

// RemoveFile safely removes a file.
func RemoveFile(path string) error {
	return os.Remove(path)
}

// CreateProgressBar creates a simple text-based progress bar.
// Clamps values to prevent panics from negative repeat counts.
func CreateProgressBar(current, total int, width int) string {
	if total <= 0 || width <= 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	filled := int(percent * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%%", bar, percent*100)
}

// TruncateString truncates a string to specified length with ellipsis.
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// GetSystemPaths returns common Windows system paths.
func GetSystemPaths() map[string]string {
	return map[string]string{
		"TEMP":         os.Getenv("TEMP"),
		"TMP":          os.Getenv("TMP"),
		"LOCALAPPDATA": os.Getenv("LOCALAPPDATA"),
		"APPDATA":      os.Getenv("APPDATA"),
		"PROGRAMDATA":  os.Getenv("PROGRAMDATA"),
		"USERPROFILE":  os.Getenv("USERPROFILE"),
		"SYSTEMROOT":   os.Getenv("SYSTEMROOT"),
		"WINDIR":       os.Getenv("WINDIR"),
	}
}

// SafeDelete attempts to delete a file or directory with retry logic for locked files.
func SafeDelete(path string, maxRetries int) error {
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error

	for i := 0; i < maxRetries; i++ {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if info.IsDir() {
			lastErr = os.RemoveAll(path)
		} else {
			lastErr = os.Remove(path)
		}

		if lastErr == nil {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("failed to delete %s after %d attempts: %w", path, maxRetries, lastErr)
}

// CleanDirectory removes files from a directory, skipping locked/protected files.
// Returns: (bytes freed, files removed, skipped count, error).
func CleanDirectory(dirPath string, maxRetries int) (int64, int, int, error) {
	var totalSize int64
	var filesRemoved int
	var filesSkipped int

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("cannot read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			size, removed, skipped, _ := CleanDirectory(fullPath, maxRetries)
			totalSize += size
			filesRemoved += removed
			filesSkipped += skipped

			_ = os.Remove(fullPath)
		} else {
			info, err := os.Stat(fullPath)
			if err != nil {
				filesSkipped++
				continue
			}
			fileSize := info.Size()

			if err := SafeDelete(fullPath, maxRetries); err != nil {
				filesSkipped++
			} else {
				totalSize += fileSize
				filesRemoved++
			}
		}
	}

	return totalSize, filesRemoved, filesSkipped, nil
}

// GetConfigDir returns the Burrow configuration directory, creating it if needed.
func GetConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	configDir := filepath.Join(appData, "Burrow")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return configDir, nil
}
