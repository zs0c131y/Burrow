package utils

import (
	"fmt"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// IsAdmin checks if the current process has administrator privileges.
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
	defer func() {
		_ = windows.FreeSid(sid)
	}()

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

// RequireAdmin returns an error if not running with admin privileges.
func RequireAdmin() error {
	if !IsAdmin() {
		return fmt.Errorf("this command requires administrator privileges; please run Burrow as Administrator")
	}
	return nil
}

// GetWindowsVersion returns Windows version information.
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
