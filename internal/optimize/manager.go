package optimize

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/zs0c131y/burrow/pkg/utils"
)

type OptimizeManager struct {
	debug  bool
	dryRun bool
}

type OptimizeTask struct {
	Name        string
	Description string
	Category    string
	Impact      string // High, Medium, Low
	Action      func() error
}

type OptimizeResults struct {
	Total          int
	Successful     int
	Failed         int
	CompletedTasks []string
	FailedTasks    []string
}

func NewOptimizeManager(debug, dryRun bool) *OptimizeManager {
	return &OptimizeManager{
		debug:  debug,
		dryRun: dryRun,
	}
}

func (om *OptimizeManager) AnalyzeSystem() ([]*OptimizeTask, error) {
	var tasks []*OptimizeTask

	// DNS Cache optimization
	tasks = append(tasks, &OptimizeTask{
		Name:        "clear_dns",
		Description: "Clear DNS cache to resolve connectivity issues",
		Category:    "Network",
		Impact:      "Low",
		Action: func() error {
			return om.clearDNSCache()
		},
	})

	// Windows Search Index
	tasks = append(tasks, &OptimizeTask{
		Name:        "rebuild_search",
		Description: "Rebuild Windows Search index for faster searches",
		Category:    "Performance",
		Impact:      "Medium",
		Action: func() error {
			return om.rebuildSearchIndex()
		},
	})

	// Icon Cache
	tasks = append(tasks, &OptimizeTask{
		Name:        "clear_icon_cache",
		Description: "Clear icon cache to fix display issues",
		Category:    "UI",
		Impact:      "Low",
		Action: func() error {
			return om.clearIconCache()
		},
	})

	// Network Reset
	tasks = append(tasks, &OptimizeTask{
		Name:        "reset_network",
		Description: "Reset network adapters and release/renew IP",
		Category:    "Network",
		Impact:      "Medium",
		Action: func() error {
			return om.resetNetwork()
		},
	})

	// Windows Update cleanup (already downloaded)
	tasks = append(tasks, &OptimizeTask{
		Name:        "cleanup_updates",
		Description: "Run Windows Update cleanup",
		Category:    "Storage",
		Impact:      "High",
		Action: func() error {
			return om.cleanupWindowsUpdate()
		},
	})

	// SFC scan (check system files)
	tasks = append(tasks, &OptimizeTask{
		Name:        "check_system_files",
		Description: "Verify system file integrity with SFC",
		Category:    "Health",
		Impact:      "High",
		Action: func() error {
			return om.runSFC()
		},
	})

	// Disable telemetry services (optional)
	tasks = append(tasks, &OptimizeTask{
		Name:        "optimize_telemetry",
		Description: "Reduce telemetry data collection",
		Category:    "Privacy",
		Impact:      "Medium",
		Action: func() error {
			return om.optimizeTelemetry()
		},
	})

	return tasks, nil
}

func (om *OptimizeManager) ExecuteOptimization(tasks []*OptimizeTask) *OptimizeResults {
	results := &OptimizeResults{
		Total: len(tasks),
	}

	for i, task := range tasks {
		progress := utils.CreateProgressBar(i+1, len(tasks), 30)
		fmt.Printf("\r%s Processing: %s", progress, utils.TruncateString(task.Description, 35))

		if om.dryRun {
			results.Successful++
			results.CompletedTasks = append(results.CompletedTasks, task.Description)
			continue
		}

		if err := task.Action(); err != nil {
			results.Failed++
			results.FailedTasks = append(results.FailedTasks, task.Description)
			if om.debug {
				color.Red("\nFailed: %s - %v", task.Name, err)
			}
		} else {
			results.Successful++
			results.CompletedTasks = append(results.CompletedTasks, task.Description)
		}
	}

	fmt.Println()
	return results
}

func (om *OptimizeManager) clearDNSCache() error {
	cmd := exec.Command("ipconfig", "/flushdns")
	return cmd.Run()
}

func (om *OptimizeManager) rebuildSearchIndex() error {
	// Stop Windows Search service
	stopCmd := exec.Command("net", "stop", "WSearch")
	if err := stopCmd.Run(); err != nil {
		return err
	}

	// Start Windows Search service
	startCmd := exec.Command("net", "start", "WSearch")
	return startCmd.Run()
}

func (om *OptimizeManager) clearIconCache() error {
	sysPaths := utils.GetSystemPaths()
	iconCachePath := sysPaths["LOCALAPPDATA"] + "\\IconCache.db"

	return utils.SafeDelete(iconCachePath, 3)
}

func (om *OptimizeManager) resetNetwork() error {
	commands := [][]string{
		{"netsh", "winsock", "reset"},
		{"netsh", "int", "ip", "reset"},
		{"ipconfig", "/release"},
		{"ipconfig", "/renew"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (om *OptimizeManager) cleanupWindowsUpdate() error {
	// Run DISM cleanup
	cmd := exec.Command("Dism.exe", "/online", "/Cleanup-Image", "/StartComponentCleanup", "/ResetBase")
	return cmd.Run()
}

func (om *OptimizeManager) runSFC() error {
	// Note: This can take a long time
	cmd := exec.Command("sfc", "/scannow")
	return cmd.Run()
}

func (om *OptimizeManager) optimizeTelemetry() error {
	// Disable DiagTrack service
	services := []string{"DiagTrack", "dmwappushservice"}

	for _, service := range services {
		cmd := exec.Command("sc", "config", service, "start=", "disabled")
		if err := cmd.Run(); err != nil {
			continue
		}

		stopCmd := exec.Command("sc", "stop", service)
		stopCmd.Run()
	}

	return nil
}
