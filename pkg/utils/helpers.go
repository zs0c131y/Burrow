package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatDuration converts duration to human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// IsAdmin checks if the current process has administrator privileges
func IsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

// RequireAdmin exits if not running with admin privileges
func RequireAdmin() {
	if !IsAdmin() {
		fmt.Println("This command requires administrator privileges.")
		fmt.Println("Please run Burrow as Administrator.")
		os.Exit(1)
	}
}

// GetDirSize calculates total size of a directory recursively
func GetDirSize(path string) (int64, int, error) {
	var size int64
	var count int

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	return size, count, err
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExpandEnvPath expands environment variables in a path
func ExpandEnvPath(path string) string {
	if strings.Contains(path, "%") {
		return os.ExpandEnv(path)
	}
	return path
}

// RemoveDirectory safely removes a directory and its contents
func RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}

// RemoveFile safely removes a file
func RemoveFile(path string) error {
	return os.Remove(path)
}

// CreateProgressBar creates a simple text-based progress bar
func CreateProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%%", bar, percent*100)
}

// TruncateString truncates a string to specified length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// GetWindowsVersion returns Windows version information
func GetWindowsVersion() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.READ)
	if err != nil {
		return "Unknown"
	}
	defer k.Close()

	product, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return "Unknown"
	}

	build, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		return product
	}

	return fmt.Sprintf("%s (Build %s)", product, build)
}

// GetSystemPaths returns common Windows system paths
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

// SafeDelete attempts to delete with retry logic for locked files
func SafeDelete(path string, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil
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

	return lastErr
}

// CleanDirectory removes files from a directory, skipping locked/protected files
// Returns: (bytes freed, files removed, skipped count, error)
func CleanDirectory(dirPath string, maxRetries int) (int64, int, int, error) {
	var totalSize int64
	var filesRemoved int
	var filesSkipped int

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, 0, 0, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Recursively clean subdirectories
			size, removed, skipped, err := CleanDirectory(fullPath, maxRetries)
			totalSize += size
			filesRemoved += removed
			filesSkipped += skipped

			// Try to remove empty directory
			if err == nil {
				os.Remove(fullPath) // Ignore error if not empty
			}
		} else {
			// Get file size before deletion
			info, err := os.Stat(fullPath)
			if err != nil {
				filesSkipped++
				continue
			}
			fileSize := info.Size()

			// Try to delete file with retries
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
