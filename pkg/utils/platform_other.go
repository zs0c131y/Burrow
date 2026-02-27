//go:build !windows

package utils

import "fmt"

// IsAdmin checks if the current process has administrator privileges.
// On non-Windows platforms, this always returns false.
func IsAdmin() bool {
	return false
}

// RequireAdmin returns an error on non-Windows platforms.
func RequireAdmin() error {
	return fmt.Errorf("this command requires Windows with administrator privileges")
}

// GetWindowsVersion returns Windows version information.
// On non-Windows platforms, this returns "Not Windows".
func GetWindowsVersion() string {
	return "Not Windows"
}
