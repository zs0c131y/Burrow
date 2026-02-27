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

// UninstallManager handles application discovery and removal.
type UninstallManager struct {
	debug  bool
	dryRun bool
}

// UninstallResult captures the outcome of an uninstall operation.
type UninstallResult struct {
	App                 *models.Application
	Success             bool
	FilesRemoved        int
	RegistryKeysRemoved int
	SpaceFreed          int64
	LocationsCleaned    []string
	Errors              []string
	Error               error
	Duration            time.Duration
}

// NewUninstallManager creates a new UninstallManager.
func NewUninstallManager(debug, dryRun bool) *UninstallManager {
	return &UninstallManager{
		debug:  debug,
		dryRun: dryRun,
	}
}

// registrySource tracks which root and path an application was discovered from,
// so we can correctly remove it later.
type registrySource struct {
	Root registry.Key
	Path string
}

// appSources maps registry key to its source for accurate cleanup.
var appSources = make(map[string]registrySource)

// DiscoverApplications scans the registry for installed applications.
func (um *UninstallManager) DiscoverApplications() ([]*models.Application, error) {
	var apps []*models.Application

	type regEntry struct {
		root registry.Key
		path string
	}

	entries := []regEntry{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
	}

	seen := make(map[string]bool)

	for _, entry := range entries {
		found, err := um.readUninstallRegistry(entry.root, entry.path)
		if err != nil {
			if um.debug {
				color.Yellow("  Warning: could not read registry at %s: %v", entry.path, err)
			}
			continue
		}
		for _, app := range found {
			key := strings.ToLower(app.DisplayName + "|" + app.Version)
			if seen[key] {
				continue
			}
			seen[key] = true
			appSources[app.RegistryKey] = registrySource{Root: entry.root, Path: entry.path}
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func (um *UninstallManager) readUninstallRegistry(root registry.Key, path string) ([]*models.Application, error) {
	var apps []*models.Application

	k, err := registry.OpenKey(root, path, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("cannot open registry key %s: %w", path, err)
	}
	defer k.Close()

	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("cannot read subkeys of %s: %w", path, err)
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
	fullPath := basePath + `\` + subkey

	k, err := registry.OpenKey(root, fullPath, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()

	app := &models.Application{
		RegistryKey: fullPath,
	}

	if val, _, err := k.GetStringValue("DisplayName"); err == nil {
		app.DisplayName = strings.TrimSpace(val)
		app.Name = val
	}

	if val, _, err := k.GetStringValue("Publisher"); err == nil {
		app.Publisher = strings.TrimSpace(val)
	}

	if val, _, err := k.GetStringValue("DisplayVersion"); err == nil {
		app.Version = strings.TrimSpace(val)
	}

	if val, _, err := k.GetStringValue("InstallLocation"); err == nil {
		app.InstallLocation = strings.TrimSpace(val)
	}

	if val, _, err := k.GetStringValue("UninstallString"); err == nil {
		app.UninstallString = strings.TrimSpace(val)
	}

	if val, _, err := k.GetStringValue("InstallDate"); err == nil {
		app.InstallDate = strings.TrimSpace(val)
	}

	if val, _, err := k.GetIntegerValue("EstimatedSize"); err == nil {
		app.Size = int64(val) * 1024
	}

	return app
}

// PreviewUninstall shows what would be removed without making changes.
func (um *UninstallManager) PreviewUninstall(app *models.Application) {
	color.White("Preview of items to be removed:\n")

	locations := um.findRelatedLocations(app)

	for _, loc := range locations {
		if utils.PathExists(loc) {
			size, count, err := utils.GetDirSize(loc)
			if err != nil {
				color.Yellow("  ! %s (cannot determine size: %v)", loc, err)
			} else {
				color.Cyan("  * %s (%s, %d files)", loc, utils.FormatBytes(size), count)
			}
		} else {
			color.Yellow("  - %s (not found)", loc)
		}
	}

	color.White("\nRegistry keys to be removed:")
	color.Cyan("  * %s", app.RegistryKey)
}

// UninstallApplication performs the full uninstall with leftover cleanup.
func (um *UninstallManager) UninstallApplication(app *models.Application) *UninstallResult {
	startTime := time.Now()

	result := &UninstallResult{
		App:     app,
		Success: true,
	}

	if um.dryRun {
		return result
	}

	// Step 1: Run native uninstaller
	color.White("Running native uninstaller...\n")
	if err := um.runNativeUninstaller(app); err != nil {
		msg := fmt.Sprintf("Native uninstaller: %v", err)
		result.Errors = append(result.Errors, msg)
		if um.debug {
			color.Yellow("  Warning: %s", msg)
		}
	}

	// Step 2: Remove leftover files
	color.White("Cleaning leftover files...\n")
	locations := um.findRelatedLocations(app)

	for _, loc := range locations {
		if !utils.PathExists(loc) {
			continue
		}
		size, count, err := utils.GetDirSize(loc)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Cannot measure %s: %v", loc, err))
			continue
		}
		if err := utils.SafeDelete(loc, 3); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Cannot remove %s: %v", loc, err))
		} else {
			result.FilesRemoved += count
			result.SpaceFreed += size
			result.LocationsCleaned = append(result.LocationsCleaned, loc)
		}
	}

	// Step 3: Remove registry entries
	color.White("Cleaning registry entries...\n")
	if err := um.removeRegistryEntries(app); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Registry cleanup: %v", err))
	} else {
		result.RegistryKeysRemoved++
	}

	result.Duration = time.Since(startTime)
	if len(result.Errors) > 0 && result.FilesRemoved == 0 && result.RegistryKeysRemoved == 0 {
		result.Success = false
		result.Error = fmt.Errorf("uninstall completed with errors")
	}

	return result
}

func (um *UninstallManager) runNativeUninstaller(app *models.Application) error {
	uninstallStr := strings.TrimSpace(app.UninstallString)
	if uninstallStr == "" {
		return fmt.Errorf("no uninstall command available")
	}

	// Handle MsiExec uninstallers
	if strings.Contains(strings.ToLower(uninstallStr), "msiexec") {
		if startIdx := strings.Index(uninstallStr, "{"); startIdx != -1 {
			if endIdx := strings.Index(uninstallStr, "}"); endIdx > startIdx {
				productCode := uninstallStr[startIdx : endIdx+1]
				cmd := exec.Command("msiexec.exe", "/x", productCode, "/qn", "/norestart")
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("msiexec uninstall failed for %s: %w", productCode, err)
				}
				return nil
			}
		}
		return fmt.Errorf("malformed MSI uninstall string: %s", uninstallStr)
	}

	// For other uninstallers, try common silent flags
	executable := strings.Trim(uninstallStr, `"`)

	// Validate the executable path exists before running
	if !utils.PathExists(executable) {
		return fmt.Errorf("uninstaller not found at: %s", executable)
	}

	cmd := exec.Command(executable, "/S", "/VERYSILENT", "/SILENT", "/NORESTART")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("uninstaller failed: %w", err)
	}
	return nil
}

func (um *UninstallManager) findRelatedLocations(app *models.Application) []string {
	var locations []string

	sysPaths := utils.GetSystemPaths()
	appName := app.Name
	if appName == "" {
		appName = app.DisplayName
	}
	if appName == "" {
		return locations
	}

	if app.InstallLocation != "" && utils.PathExists(app.InstallLocation) {
		locations = append(locations, app.InstallLocation)
	}

	potentialPaths := []string{
		filepath.Join(sysPaths["LOCALAPPDATA"], appName),
		filepath.Join(sysPaths["APPDATA"], appName),
		filepath.Join(sysPaths["PROGRAMDATA"], appName),
	}

	seen := make(map[string]bool)
	if app.InstallLocation != "" {
		seen[strings.ToLower(app.InstallLocation)] = true
	}

	for _, path := range potentialPaths {
		lower := strings.ToLower(path)
		if seen[lower] {
			continue
		}
		seen[lower] = true
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

	// Determine which registry root this app was discovered from
	source, ok := appSources[app.RegistryKey]
	if !ok {
		return fmt.Errorf("cannot determine registry root for key: %s", app.RegistryKey)
	}

	// The RegistryKey is the full subpath under the root, e.g.:
	// SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\AppName
	// We need to open the parent and delete the last segment.
	lastSep := strings.LastIndex(app.RegistryKey, `\`)
	if lastSep < 0 {
		return fmt.Errorf("invalid registry key format: %s", app.RegistryKey)
	}

	parentPath := app.RegistryKey[:lastSep]
	subkeyName := app.RegistryKey[lastSep+1:]

	k, err := registry.OpenKey(source.Root, parentPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("cannot open registry key %s: %w", parentPath, err)
	}
	defer k.Close()

	if err := registry.DeleteKey(k, subkeyName); err != nil {
		return fmt.Errorf("cannot delete registry subkey %s: %w", subkeyName, err)
	}
	return nil
}
