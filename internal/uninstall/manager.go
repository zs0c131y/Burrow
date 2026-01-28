package uninstall

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/zs0c131y/burrow/pkg/models"
	"github.com/zs0c131y/burrow/pkg/utils"
	"golang.org/x/sys/windows/registry"
)

type UninstallManager struct {
	debug  bool
	dryRun bool
}

type UninstallResult struct {
	App                 *models.Application
	Success             bool
	FilesRemoved        int
	RegistryKeysRemoved int
	SpaceFreed          int64
	LocationsCleaned    []string
	Error               error
	Duration            time.Duration
}

func NewUninstallManager(debug, dryRun bool) *UninstallManager {
	return &UninstallManager{
		debug:  debug,
		dryRun: dryRun,
	}
}

func (um *UninstallManager) DiscoverApplications() ([]*models.Application, error) {
	var apps []*models.Application

	// Check both 64-bit and 32-bit registry locations
	registryPaths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, regPath := range registryPaths {
		appsFromReg, err := um.readUninstallRegistry(registry.LOCAL_MACHINE, regPath)
		if err == nil {
			apps = append(apps, appsFromReg...)
		}
	}

	// Also check current user registry
	userApps, err := um.readUninstallRegistry(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`)
	if err == nil {
		apps = append(apps, userApps...)
	}

	return apps, nil
}

func (um *UninstallManager) readUninstallRegistry(root registry.Key, path string) ([]*models.Application, error) {
	var apps []*models.Application

	k, err := registry.OpenKey(root, path, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return apps, err
	}
	defer k.Close()

	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return apps, err
	}

	for _, subkey := range subkeys {
		app := um.readApplicationInfo(root, path, subkey)
		if app != nil && app.DisplayName != "" && app.UninstallString != "" {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func (um *UninstallManager) readApplicationInfo(root registry.Key, basePath, subkey string) *models.Application {
	fullPath := filepath.Join(basePath, subkey)

	k, err := registry.OpenKey(root, fullPath, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()

	app := &models.Application{
		RegistryKey: fullPath,
	}

	// Read string values
	if val, _, err := k.GetStringValue("DisplayName"); err == nil {
		app.DisplayName = val
		app.Name = val
	}

	if val, _, err := k.GetStringValue("Publisher"); err == nil {
		app.Publisher = val
	}

	if val, _, err := k.GetStringValue("DisplayVersion"); err == nil {
		app.Version = val
	}

	if val, _, err := k.GetStringValue("InstallLocation"); err == nil {
		app.InstallLocation = val
	}

	if val, _, err := k.GetStringValue("UninstallString"); err == nil {
		app.UninstallString = val
	}

	if val, _, err := k.GetStringValue("InstallDate"); err == nil {
		app.InstallDate = val
	}

	// Read size (DWORD)
	if val, _, err := k.GetIntegerValue("EstimatedSize"); err == nil {
		app.Size = int64(val) * 1024 // Size is in KB
	}

	return app
}

func (um *UninstallManager) PreviewUninstall(app *models.Application) {
	color.White("Preview of items to be removed:\n")

	locations := um.findRelatedLocations(app)

	for _, loc := range locations {
		if utils.PathExists(loc) {
			size, count, _ := utils.GetDirSize(loc)
			color.Cyan("  • %s (%s, %d files)", loc, utils.FormatBytes(size), count)
		} else {
			color.Yellow("  • %s (not found)", loc)
		}
	}

	color.White("\nRegistry keys to be removed:")
	color.Cyan("  • %s", app.RegistryKey)
}

func (um *UninstallManager) UninstallApplication(app *models.Application) *UninstallResult {
	startTime := time.Now()

	result := &UninstallResult{
		App:     app,
		Success: true,
	}

	if um.dryRun {
		result.Success = true
		return result
	}

	// Step 1: Run native uninstaller
	color.White("Running native uninstaller...\n")
	if err := um.runNativeUninstaller(app); err != nil {
		if um.debug {
			color.Yellow("Warning: Native uninstaller failed: %v", err)
		}
	}

	// Step 2: Remove leftover files
	color.White("Cleaning leftover files...\n")
	locations := um.findRelatedLocations(app)

	for _, loc := range locations {
		if utils.PathExists(loc) {
			size, count, _ := utils.GetDirSize(loc)
			if err := utils.SafeDelete(loc, 3); err == nil {
				result.FilesRemoved += count
				result.SpaceFreed += size
				result.LocationsCleaned = append(result.LocationsCleaned, loc)
			}
		}
	}

	// Step 3: Remove registry entries
	color.White("Cleaning registry entries...\n")
	if err := um.removeRegistryEntries(app); err == nil {
		result.RegistryKeysRemoved++
	}

	result.Duration = time.Since(startTime)
	return result
}

func (um *UninstallManager) runNativeUninstaller(app *models.Application) error {
	uninstallStr := app.UninstallString

	// Handle MsiExec uninstallers
	if strings.Contains(strings.ToLower(uninstallStr), "msiexec") {
		// Extract product code and run silent uninstall
		if strings.Contains(uninstallStr, "{") {
			startIdx := strings.Index(uninstallStr, "{")
			endIdx := strings.Index(uninstallStr, "}")
			if startIdx != -1 && endIdx != -1 {
				productCode := uninstallStr[startIdx : endIdx+1]
				cmd := exec.Command("msiexec.exe", "/x", productCode, "/qn", "/norestart")
				return cmd.Run()
			}
		}
	}

	// For other uninstallers, try to run with silent flags
	uninstallStr = strings.Trim(uninstallStr, "\"")
	cmd := exec.Command(uninstallStr, "/S", "/VERYSILENT", "/SILENT")
	return cmd.Run()
}

func (um *UninstallManager) findRelatedLocations(app *models.Application) []string {
	var locations []string

	sysPaths := utils.GetSystemPaths()
	appName := app.Name

	// Install location
	if app.InstallLocation != "" {
		locations = append(locations, app.InstallLocation)
	}

	// Common locations for app data
	potentialPaths := []string{
		filepath.Join(sysPaths["LOCALAPPDATA"], appName),
		filepath.Join(sysPaths["APPDATA"], appName),
		filepath.Join(sysPaths["PROGRAMDATA"], appName),
		filepath.Join(sysPaths["USERPROFILE"], "AppData", "Local", appName),
		filepath.Join(sysPaths["USERPROFILE"], "AppData", "Roaming", appName),
	}

	for _, path := range potentialPaths {
		if utils.PathExists(path) {
			locations = append(locations, path)
		}
	}

	return locations
}

func (um *UninstallManager) removeRegistryEntries(app *models.Application) error {
	if app.RegistryKey == "" {
		return nil
	}

	// Determine which registry root
	var root registry.Key
	if strings.Contains(app.RegistryKey, "CURRENT_USER") {
		root = registry.CURRENT_USER
	} else {
		root = registry.LOCAL_MACHINE
	}

	// Extract path
	parts := strings.Split(app.RegistryKey, "\\")
	if len(parts) < 2 {
		return fmt.Errorf("invalid registry key format")
	}

	keyPath := strings.Join(parts[1:len(parts)-1], "\\")
	subkey := parts[len(parts)-1]

	k, err := registry.OpenKey(root, keyPath, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()

	return registry.DeleteKey(k, subkey)
}
