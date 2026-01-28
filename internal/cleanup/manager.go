package cleanup

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/zs0c131y/burrow/pkg/models"
	"github.com/zs0c131y/burrow/pkg/utils"
)

type CleanupManager struct {
	debug     bool
	dryRun    bool
	whitelist map[string]bool
	mutex     sync.Mutex
}

type CleanupSummary struct {
	TotalTargets      int
	SuccessfulCleans  int
	FailedCleans      int
	TotalSpaceFreed   int64
	TotalFilesRemoved int
	Results           []*CleanupResult
}

type CleanupResult struct {
	Target       *models.CleanupTarget
	Success      bool
	SpaceFreed   int64
	FilesRemoved int
	Error        error
}

func NewCleanupManager(debug, dryRun bool) *CleanupManager {
	return &CleanupManager{
		debug:     debug,
		dryRun:    dryRun,
		whitelist: loadWhitelist(),
	}
}

func (cm *CleanupManager) DiscoverTargets(categories []string) ([]*models.CleanupTarget, error) {
	var targets []*models.CleanupTarget

	sysPaths := utils.GetSystemPaths()

	// Temporary Files
	if len(categories) == 0 || contains(categories, "temp") {
		targets = append(targets, cm.getTempTargets(sysPaths)...)
	}

	// Cache Files
	if len(categories) == 0 || contains(categories, "cache") {
		targets = append(targets, cm.getCacheTargets(sysPaths)...)
	}

	// Browser Data
	if len(categories) == 0 || contains(categories, "browser") {
		targets = append(targets, cm.getBrowserTargets(sysPaths)...)
	}

	// Windows Update
	if len(categories) == 0 || contains(categories, "updates") {
		targets = append(targets, cm.getWindowsUpdateTargets(sysPaths)...)
	}

	// Logs
	if len(categories) == 0 || contains(categories, "logs") {
		targets = append(targets, cm.getLogTargets(sysPaths)...)
	}

	// Other cleanup targets
	targets = append(targets, cm.getOtherTargets(sysPaths)...)

	// Calculate sizes
	for _, target := range targets {
		if utils.PathExists(target.Path) {
			size, count, _ := utils.GetDirSize(target.Path)
			target.Size = size
			target.ItemCount = count
			target.Protected = cm.isProtected(target.Path)
		}
	}

	// Filter out empty targets
	var nonEmptyTargets []*models.CleanupTarget
	for _, target := range targets {
		if target.Size > 0 {
			nonEmptyTargets = append(nonEmptyTargets, target)
		}
	}

	return nonEmptyTargets, nil
}

func (cm *CleanupManager) getTempTargets(paths map[string]string) []*models.CleanupTarget {
	return []*models.CleanupTarget{
		{
			Name:        "Windows Temp",
			Path:        filepath.Join(paths["WINDIR"], "Temp"),
			Description: "System temporary files",
			Category:    models.CategoryTemp,
		},
		{
			Name:        "User Temp",
			Path:        paths["TEMP"],
			Description: "User temporary files",
			Category:    models.CategoryTemp,
		},
		{
			Name:        "Local Temp",
			Path:        paths["TMP"],
			Description: "Local temporary storage",
			Category:    models.CategoryTemp,
		},
	}
}

func (cm *CleanupManager) getCacheTargets(paths map[string]string) []*models.CleanupTarget {
	targets := []*models.CleanupTarget{
		{
			Name:        "Application Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "cache"),
			Description: "Application cache files",
			Category:    models.CategoryCache,
		},
		{
			Name:        "Icon Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "IconCache.db"),
			Description: "Windows icon cache",
			Category:    models.CategoryThumbnails,
		},
		{
			Name:        "Thumbnail Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "Microsoft", "Windows", "Explorer"),
			Description: "Windows thumbnail cache",
			Category:    models.CategoryThumbnails,
		},
		{
			Name:        "Prefetch",
			Path:        filepath.Join(paths["WINDIR"], "Prefetch"),
			Description: "Windows prefetch files",
			Category:    models.CategoryPrefetch,
		},
	}

	return targets
}

func (cm *CleanupManager) getBrowserTargets(paths map[string]string) []*models.CleanupTarget {
	targets := []*models.CleanupTarget{
		{
			Name:        "Chrome Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "Google", "Chrome", "User Data", "Default", "Cache"),
			Description: "Google Chrome cache",
			Category:    models.CategoryBrowser,
		},
		{
			Name:        "Edge Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "Microsoft", "Edge", "User Data", "Default", "Cache"),
			Description: "Microsoft Edge cache",
			Category:    models.CategoryBrowser,
		},
		{
			Name:        "Firefox Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "Mozilla", "Firefox", "Profiles"),
			Description: "Mozilla Firefox cache",
			Category:    models.CategoryBrowser,
		},
		{
			Name:        "Brave Cache",
			Path:        filepath.Join(paths["LOCALAPPDATA"], "BraveSoftware", "Brave-Browser", "User Data", "Default", "Cache"),
			Description: "Brave browser cache",
			Category:    models.CategoryBrowser,
		},
	}

	return targets
}

func (cm *CleanupManager) getWindowsUpdateTargets(paths map[string]string) []*models.CleanupTarget {
	return []*models.CleanupTarget{
		{
			Name:        "Windows Update Cache",
			Path:        filepath.Join(paths["WINDIR"], "SoftwareDistribution", "Download"),
			Description: "Windows Update downloaded files",
			Category:    models.CategoryWindowsUpdate,
		},
		{
			Name:        "Delivery Optimization",
			Path:        filepath.Join(paths["WINDIR"], "ServiceProfiles", "NetworkService", "AppData", "Local", "Microsoft", "Windows", "DeliveryOptimization", "Cache"),
			Description: "Windows Update delivery optimization",
			Category:    models.CategoryWindowsUpdate,
		},
	}
}

func (cm *CleanupManager) getLogTargets(paths map[string]string) []*models.CleanupTarget {
	return []*models.CleanupTarget{
		{
			Name:        "Windows Logs",
			Path:        filepath.Join(paths["WINDIR"], "Logs"),
			Description: "Windows system logs",
			Category:    models.CategoryLogs,
		},
		{
			Name:        "CBS Logs",
			Path:        filepath.Join(paths["WINDIR"], "Logs", "CBS"),
			Description: "Component-Based Servicing logs",
			Category:    models.CategoryLogs,
		},
		{
			Name:        "Panther Logs",
			Path:        filepath.Join(paths["WINDIR"], "Panther"),
			Description: "Windows installation logs",
			Category:    models.CategoryLogs,
		},
	}
}

func (cm *CleanupManager) getOtherTargets(paths map[string]string) []*models.CleanupTarget {
	return []*models.CleanupTarget{
		{
			Name:        "Recycle Bin",
			Path:        filepath.Join(paths["SYSTEMROOT"], "$Recycle.Bin"),
			Description: "Recycle bin contents",
			Category:    models.CategoryRecycleBin,
		},
		{
			Name:        "Windows Error Reporting",
			Path:        filepath.Join(paths["PROGRAMDATA"], "Microsoft", "Windows", "WER"),
			Description: "Windows error reports",
			Category:    models.CategoryLogs,
		},
	}
}

func (cm *CleanupManager) ExecuteCleanup(targets []*models.CleanupTarget) *CleanupSummary {
	summary := &CleanupSummary{
		TotalTargets: len(targets),
		Results:      make([]*CleanupResult, 0),
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

	startSize := target.Size
	startCount := target.ItemCount

	if err := utils.SafeDelete(target.Path, 3); err != nil {
		result.Success = false
		result.Error = err
		if cm.debug {
			color.Red("\nError cleaning %s: %v", target.Name, err)
		}
	} else {
		result.SpaceFreed = startSize
		result.FilesRemoved = startCount
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
		if s == item {
			return true
		}
	}
	return false
}

func loadWhitelist() map[string]bool {
	// Load from config file in future
	return make(map[string]bool)
}

func ManageWhitelist() {
	color.Cyan("\nWhitelist Management")
	color.White("════════════════════════════════════════════════════════\n")
	color.Yellow("Feature coming soon: Manage protected paths")
	color.White("\nProtected paths will not be cleaned during cleanup operations.")
}
