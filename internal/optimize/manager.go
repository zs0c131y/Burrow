package optimize

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/zs0c131y/burrow/pkg/utils"
)

// OptimizeManager handles system optimization tasks.
type OptimizeManager struct {
	debug  bool
	dryRun bool
}

// OptimizeTask represents a single optimization action.
type OptimizeTask struct {
	Name        string
	Description string
	Category    string
	Impact      string // High, Medium, Low
	Action      func() error
}

// OptimizeResults captures the outcome of an optimization run.
type OptimizeResults struct {
	Total          int
	Successful     int
	Failed         int
	CompletedTasks []string
	FailedTasks    []string
}

// NewOptimizeManager creates a new OptimizeManager.
func NewOptimizeManager(debug, dryRun bool) *OptimizeManager {
	return &OptimizeManager{
		debug:  debug,
		dryRun: dryRun,
	}
}

// AnalyzeSystem returns the list of available optimization tasks.
func (om *OptimizeManager) AnalyzeSystem() ([]*OptimizeTask, error) {
	tasks := []*OptimizeTask{
		{
			Name:        "clear_dns",
			Description: "Clear DNS cache to resolve connectivity issues",
			Category:    "Network",
			Impact:      "Low",
			Action:      om.clearDNSCache,
		},
		{
			Name:        "rebuild_search",
			Description: "Rebuild Windows Search index for faster searches",
			Category:    "Performance",
			Impact:      "Medium",
			Action:      om.rebuildSearchIndex,
		},
		{
			Name:        "clear_icon_cache",
			Description: "Clear icon cache to fix display issues",
			Category:    "UI",
			Impact:      "Low",
			Action:      om.clearIconCache,
		},
		{
			Name:        "reset_network",
			Description: "Reset network adapters and release/renew IP",
			Category:    "Network",
			Impact:      "Medium",
			Action:      om.resetNetwork,
		},
		{
			Name:        "cleanup_updates",
			Description: "Run Windows Update cleanup",
			Category:    "Storage",
			Impact:      "High",
			Action:      om.cleanupWindowsUpdate,
		},
		{
			Name:        "check_system_files",
			Description: "Verify system file integrity with SFC",
			Category:    "Health",
			Impact:      "High",
			Action:      om.runSFC,
		},
		{
			Name:        "optimize_telemetry",
			Description: "Reduce telemetry data collection",
			Category:    "Privacy",
			Impact:      "Medium",
			Action:      om.optimizeTelemetry,
		},
	}

	return tasks, nil
}

// ExecuteOptimization runs all provided tasks and returns results.
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
			errMsg := task.Description
			if om.debug {
				errMsg = fmt.Sprintf("%s (%v)", task.Description, err)
			}
			results.FailedTasks = append(results.FailedTasks, errMsg)
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipconfig /flushdns failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (om *OptimizeManager) rebuildSearchIndex() error {
	stopCmd := exec.Command("net", "stop", "WSearch")
	if output, err := stopCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop WSearch service: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	startCmd := exec.Command("net", "start", "WSearch")
	if output, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start WSearch service: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func (om *OptimizeManager) clearIconCache() error {
	sysPaths := utils.GetSystemPaths()
	localAppData := sysPaths["LOCALAPPDATA"]
	if localAppData == "" {
		return fmt.Errorf("LOCALAPPDATA environment variable not set")
	}

	iconCachePath := localAppData + `\IconCache.db`

	if !utils.PathExists(iconCachePath) {
		return nil
	}

	return utils.SafeDelete(iconCachePath, 3)
}

func (om *OptimizeManager) resetNetwork() error {
	type netCmd struct {
		args    []string
		canFail bool
	}

	commands := []netCmd{
		{args: []string{"ipconfig", "/release"}, canFail: true},
		{args: []string{"ipconfig", "/flushdns"}, canFail: true},
		{args: []string{"netsh", "winsock", "reset"}, canFail: false},
		{args: []string{"netsh", "int", "ip", "reset"}, canFail: false},
		{args: []string{"ipconfig", "/renew"}, canFail: false},
	}

	var errors []string

	for _, nc := range commands {
		cmd := exec.Command(nc.args[0], nc.args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			msg := fmt.Sprintf("%s: %v (output: %s)",
				strings.Join(nc.args, " "),
				err,
				strings.TrimSpace(string(output)))
			if !nc.canFail {
				errors = append(errors, msg)
			} else if om.debug {
				color.Yellow("  Warning: %s", msg)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("network reset had errors: %s", strings.Join(errors, "; "))
	}
	return nil
}

func (om *OptimizeManager) cleanupWindowsUpdate() error {
	cmd := exec.Command("Dism.exe", "/online", "/Cleanup-Image", "/StartComponentCleanup", "/ResetBase")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("DISM cleanup failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (om *OptimizeManager) runSFC() error {
	cmd := exec.Command("sfc", "/scannow")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("SFC scan failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (om *OptimizeManager) optimizeTelemetry() error {
	services := []string{"DiagTrack", "dmwappushservice"}
	var errors []string

	for _, service := range services {
		configCmd := exec.Command("sc", "config", service, "start=", "disabled")
		if output, err := configCmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("sc config %s: %v (%s)",
				service, err, strings.TrimSpace(string(output))))
			continue
		}

		stopCmd := exec.Command("sc", "stop", service)
		if output, err := stopCmd.CombinedOutput(); err != nil {
			if om.debug {
				color.Yellow("  Warning: could not stop %s: %v (%s)",
					service, err, strings.TrimSpace(string(output)))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("telemetry optimization had errors: %s", strings.Join(errors, "; "))
	}
	return nil
}
