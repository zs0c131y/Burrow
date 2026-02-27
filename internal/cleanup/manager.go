package cleanup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/zs0c131y/burrow/pkg/models"
	"github.com/zs0c131y/burrow/pkg/utils"
)

// CleanupManager handles system cleanup operations.
type CleanupManager struct {
	debug     bool
	dryRun    bool
	whitelist map[string]bool
	mutex     sync.Mutex
}

// CleanupSummary captures the results of a cleanup run.
type CleanupSummary struct {
	TotalTargets      int
	SuccessfulCleans  int
	FailedCleans      int
	TotalSpaceFreed   int64
	TotalFilesRemoved int
	Results           []*CleanupResult
}

// CleanupResult captures the result of cleaning a single target.
type CleanupResult struct {
	Target       *models.CleanupTarget
	Success      bool
	SpaceFreed   int64
	FilesRemoved int
	Error        error
}

// NewCleanupManager creates a new CleanupManager.
func NewCleanupManager(debug, dryRun bool) *CleanupManager {
	return &CleanupManager{
		debug:     debug,
		dryRun:    dryRun,
		whitelist: loadWhitelist(),
	}
}

// DiscoverTargets finds cleanup targets based on the given category filter.
func (cm *CleanupManager) DiscoverTargets(categories []string) ([]*models.CleanupTarget, error) {
	var targets []*models.CleanupTarget

	sysPaths := utils.GetSystemPaths()

	if len(categories) == 0 || contains(categories, "temp") {
		targets = append(targets, cm.getTempTargets(sysPaths)...)
	}

	if len(categories) == 0 || contains(categories, "cache") {
		targets = append(targets, cm.getCacheTargets(sysPaths)...)
	}

	if len(categories) == 0 || contains(categories, "browser") {
		targets = append(targets, cm.getBrowserTargets(sysPaths)...)
	}

	if len(categories) == 0 || contains(categories, "updates") {
		targets = append(targets, cm.getWindowsUpdateTargets(sysPaths)...)
	}

	if len(categories) == 0 || contains(categories, "logs") {
		targets = append(targets, cm.getLogTargets(sysPaths)...)
	}

	targets = append(targets, cm.getOtherTargets(sysPaths)...)

	for _, target := range targets {
		if utils.PathExists(target.Path) {
			size, count, err := utils.GetDirSize(target.Path)
			if err != nil && cm.debug {
				color.Yellow("  Warning: error scanning %s: %v", target.Name, err)
			}
			target.Size = size
			target.ItemCount = count
			target.Protected = cm.isProtected(target.Path)
		}
	}

	var nonEmptyTargets []*models.CleanupTarget
	for _, target := range targets {
		if target.Size > 0 {
			nonEmptyTargets = append(nonEmptyTargets, target)
		}
	}

	return nonEmptyTargets, nil
}

func (cm *CleanupManager) getTempTargets(paths map[string]string) []*models.CleanupTarget {
	var targets []*models.CleanupTarget

	if windir := paths["WINDIR"]; windir != "" {
		targets = append(targets, &models.CleanupTarget{
			Name:        "Windows Temp",
			Path:        filepath.Join(windir, "Temp"),
			Description: "System temporary files",
			Category:    models.CategoryTemp,
		})
	}

	if temp := paths["TEMP"]; temp != "" {
		targets = append(targets, &models.CleanupTarget{
			Name:        "User Temp",
			Path:        temp,
			Description: "User temporary files",
			Category:    models.CategoryTemp,
		})
	}

	if tmp := paths["TMP"]; tmp != "" && tmp != paths["TEMP"] {
		targets = append(targets, &models.CleanupTarget{
			Name:        "Local Temp",
			Path:        tmp,
			Description: "Local temporary storage",
			Category:    models.CategoryTemp,
		})
	}

	return targets
}

func (cm *CleanupManager) getCacheTargets(paths map[string]string) []*models.CleanupTarget {
	var targets []*models.CleanupTarget
	localAppData := paths["LOCALAPPDATA"]
	windir := paths["WINDIR"]

	if localAppData != "" {
		targets = append(targets,
			&models.CleanupTarget{
				Name:        "Application Cache",
				Path:        filepath.Join(localAppData, "cache"),
				Description: "Application cache files",
				Category:    models.CategoryCache,
			},
			&models.CleanupTarget{
				Name:        "Icon Cache",
				Path:        filepath.Join(localAppData, "IconCache.db"),
				Description: "Windows icon cache",
				Category:    models.CategoryThumbnails,
			},
			&models.CleanupTarget{
				Name:        "Thumbnail Cache",
				Path:        filepath.Join(localAppData, "Microsoft", "Windows", "Explorer"),
				Description: "Windows thumbnail cache",
				Category:    models.CategoryThumbnails,
			},
		)
	}

	if windir != "" {
		targets = append(targets, &models.CleanupTarget{
			Name:        "Prefetch",
			Path:        filepath.Join(windir, "Prefetch"),
			Description: "Windows prefetch files",
			Category:    models.CategoryPrefetch,
		})
	}

	return targets
}

func (cm *CleanupManager) getBrowserTargets(paths map[string]string) []*models.CleanupTarget {
	var targets []*models.CleanupTarget
	localAppData := paths["LOCALAPPDATA"]
	if localAppData == "" {
		return targets
	}

	browsers := []struct {
		name string
		path string
	}{
		{"Chrome Cache", filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Cache")},
		{"Edge Cache", filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Cache")},
		{"Firefox Cache", filepath.Join(localAppData, "Mozilla", "Firefox", "Profiles")},
		{"Brave Cache", filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "Default", "Cache")},
	}

	for _, b := range browsers {
		targets = append(targets, &models.CleanupTarget{
			Name:        b.name,
			Path:        b.path,
			Description: b.name,
			Category:    models.CategoryBrowser,
		})
	}

	return targets
}

func (cm *CleanupManager) getWindowsUpdateTargets(paths map[string]string) []*models.CleanupTarget {
	var targets []*models.CleanupTarget
	windir := paths["WINDIR"]
	if windir == "" {
		return targets
	}

	targets = append(targets,
		&models.CleanupTarget{
			Name:        "Windows Update Cache",
			Path:        filepath.Join(windir, "SoftwareDistribution", "Download"),
			Description: "Windows Update downloaded files",
			Category:    models.CategoryWindowsUpdate,
		},
		&models.CleanupTarget{
			Name:        "Delivery Optimization",
			Path:        filepath.Join(windir, "ServiceProfiles", "NetworkService", "AppData", "Local", "Microsoft", "Windows", "DeliveryOptimization", "Cache"),
			Description: "Windows Update delivery optimization",
			Category:    models.CategoryWindowsUpdate,
		},
	)

	return targets
}

func (cm *CleanupManager) getLogTargets(paths map[string]string) []*models.CleanupTarget {
	var targets []*models.CleanupTarget
	windir := paths["WINDIR"]
	if windir == "" {
		return targets
	}

	targets = append(targets,
		&models.CleanupTarget{
			Name:        "Windows Logs",
			Path:        filepath.Join(windir, "Logs"),
			Description: "Windows system logs",
			Category:    models.CategoryLogs,
		},
		&models.CleanupTarget{
			Name:        "CBS Logs",
			Path:        filepath.Join(windir, "Logs", "CBS"),
			Description: "Component-Based Servicing logs",
			Category:    models.CategoryLogs,
		},
		&models.CleanupTarget{
			Name:        "Panther Logs",
			Path:        filepath.Join(windir, "Panther"),
			Description: "Windows installation logs",
			Category:    models.CategoryLogs,
		},
	)

	return targets
}

func (cm *CleanupManager) getOtherTargets(paths map[string]string) []*models.CleanupTarget {
	var targets []*models.CleanupTarget

	if sysRoot := paths["SYSTEMROOT"]; sysRoot != "" {
		targets = append(targets, &models.CleanupTarget{
			Name:        "Recycle Bin",
			Path:        filepath.Join(sysRoot, "$Recycle.Bin"),
			Description: "Recycle bin contents",
			Category:    models.CategoryRecycleBin,
		})
	}

	if progData := paths["PROGRAMDATA"]; progData != "" {
		targets = append(targets, &models.CleanupTarget{
			Name:        "Windows Error Reporting",
			Path:        filepath.Join(progData, "Microsoft", "Windows", "WER"),
			Description: "Windows error reports",
			Category:    models.CategoryLogs,
		})
	}

	return targets
}

// ExecuteCleanup runs the cleanup on all non-protected targets.
func (cm *CleanupManager) ExecuteCleanup(targets []*models.CleanupTarget) *CleanupSummary {
	summary := &CleanupSummary{
		TotalTargets: len(targets),
		Results:      make([]*CleanupResult, 0, len(targets)),
	}

	for i, target := range targets {
		if target.Protected {
			continue
		}

		progress := utils.CreateProgressBar(i+1, len(targets), 30)
		fmt.Printf("\r%s Processing: %s", progress, utils.TruncateString(target.Name, 35))

		result := cm.cleanTarget(target)
		summary.Results = append(summary.Results, result)

		if result.Success {
			summary.SuccessfulCleans++
			summary.TotalSpaceFreed += result.SpaceFreed
			summary.TotalFilesRemoved += result.FilesRemoved
		} else {
			summary.FailedCleans++
		}
	}

	fmt.Println()
	return summary
}

func (cm *CleanupManager) cleanTarget(target *models.CleanupTarget) *CleanupResult {
	result := &CleanupResult{
		Target:  target,
		Success: true,
	}

	if cm.dryRun {
		result.SpaceFreed = target.Size
		result.FilesRemoved = target.ItemCount
		return result
	}

	freedSpace, filesRemoved, filesSkipped, err := utils.CleanDirectory(target.Path, 3)

	if err != nil {
		result.Success = false
		result.Error = err
		if cm.debug {
			color.Red("\nError accessing %s: %v", target.Name, err)
		}
		return result
	}

	result.SpaceFreed = freedSpace
	result.FilesRemoved = filesRemoved

	if filesRemoved == 0 && filesSkipped > 0 {
		result.Success = false
		result.Error = fmt.Errorf("%d files locked or in use (skipped)", filesSkipped)
	}

	if cm.debug && filesSkipped > 0 {
		color.Yellow("\n  %s: %d files skipped (locked/in-use)", target.Name, filesSkipped)
	}

	return result
}

func (cm *CleanupManager) isProtected(path string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.whitelist[strings.ToLower(path)]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// whitelistConfig is the on-disk format for the whitelist file.
type whitelistConfig struct {
	Paths []string `json:"protected_paths"`
}

func getWhitelistPath() string {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "whitelist.json")
}

func loadWhitelist() map[string]bool {
	result := make(map[string]bool)

	path := getWhitelistPath()
	if path == "" {
		return result
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}

	var cfg whitelistConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return result
	}

	for _, p := range cfg.Paths {
		result[strings.ToLower(p)] = true
	}

	return result
}

func saveWhitelist(paths map[string]bool) error {
	wlPath := getWhitelistPath()
	if wlPath == "" {
		return fmt.Errorf("cannot determine config directory")
	}

	var pathList []string
	for p := range paths {
		pathList = append(pathList, p)
	}

	cfg := whitelistConfig{Paths: pathList}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal whitelist: %w", err)
	}

	if err := os.WriteFile(wlPath, data, 0o644); err != nil {
		return fmt.Errorf("cannot write whitelist file: %w", err)
	}

	return nil
}

// ManageWhitelist provides interactive whitelist management.
func ManageWhitelist() {
	color.Cyan("\nWhitelist Management")
	color.White("════════════════════════════════════════════════════════\n")
	color.White("Protected paths will not be cleaned during cleanup operations.\n")

	wl := loadWhitelist()

	if len(wl) > 0 {
		color.White("\nCurrently protected paths:\n")
		i := 1
		for p := range wl {
			fmt.Printf("  %d. %s\n", i, p)
			i++
		}
	} else {
		color.Yellow("\nNo paths are currently protected.\n")
	}

	fmt.Println()
	color.White("Options:\n")
	color.White("  1. Add a path to whitelist\n")
	color.White("  2. Remove a path from whitelist\n")
	color.White("  3. Exit whitelist management\n")
	fmt.Print("\nSelect option (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Enter path to protect: ")
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)
		if path == "" {
			color.Yellow("No path entered.")
			return
		}
		wl[strings.ToLower(path)] = true
		if err := saveWhitelist(wl); err != nil {
			color.Red("Error saving whitelist: %v", err)
			return
		}
		color.Green("Path added to whitelist: %s", path)

	case "2":
		fmt.Print("Enter path to remove: ")
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)
		if path == "" {
			color.Yellow("No path entered.")
			return
		}
		lower := strings.ToLower(path)
		if !wl[lower] {
			color.Yellow("Path not found in whitelist.")
			return
		}
		delete(wl, lower)
		if err := saveWhitelist(wl); err != nil {
			color.Red("Error saving whitelist: %v", err)
			return
		}
		color.Green("Path removed from whitelist: %s", path)

	case "3":
		return

	default:
		color.Yellow("Invalid option.")
	}
}
