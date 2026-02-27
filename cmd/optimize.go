package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zs0c131y/burrow/internal/optimize"
	"github.com/zs0c131y/burrow/pkg/utils"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "System optimization and tuning",
	Long: `Optimize Windows system performance:
  - Disable unnecessary startup programs
  - Stop resource-heavy services
  - Clean registry temporary entries
  - Optimize network settings
  - Clear DNS cache
  - Rebuild system databases
  - Optimize power settings`,
	Run: func(cmd *cobra.Command, args []string) {
		runOptimize()
	},
}

func runOptimize() {
	if !dryRun {
		if err := utils.RequireAdmin(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	}

	color.Cyan("\n╔════════════════════════════════════════════════════════╗")
	color.Cyan("║             Burrow System Optimization                 ║")
	color.Cyan("╚════════════════════════════════════════════════════════╝\n")

	if dryRun {
		color.Yellow("DRY RUN MODE - No changes will be made\n")
	}

	startTime := time.Now()

	manager := optimize.NewOptimizeManager(debugMode, dryRun)

	color.White("Analyzing system...\n")

	tasks, err := manager.AnalyzeSystem()
	if err != nil {
		color.Red("Error analyzing system: %v", err)
		return
	}

	if len(tasks) == 0 {
		color.Green("System is already optimized!")
		return
	}

	color.White("\nOptimization Tasks:\n")
	color.White("════════════════════════════════════════════════════════\n")

	for i, task := range tasks {
		var statusIcon string
		switch task.Impact {
		case "High":
			statusIcon = color.RedString("!")
		case "Medium":
			statusIcon = color.YellowString("~")
		default:
			statusIcon = color.GreenString("*")
		}

		fmt.Printf("  %s %d. %s\n", statusIcon, i+1, task.Description)
		fmt.Printf("     Impact: %s | Category: %s\n", task.Impact, task.Category)
	}

	color.White("\n════════════════════════════════════════════════════════\n")
	fmt.Printf("Total Tasks: %d\n", len(tasks))
	color.White("════════════════════════════════════════════════════════\n\n")

	if dryRun {
		color.Yellow("Dry run complete. No changes were made.")
		return
	}

	if !confirmAction("Proceed with optimization?") {
		color.Yellow("Optimization cancelled.")
		return
	}

	fmt.Println()
	color.White("Optimizing system...\n")

	results := manager.ExecuteOptimization(tasks)

	displayOptimizeResults(results, time.Since(startTime))
}

func displayOptimizeResults(results *optimize.OptimizeResults, duration time.Duration) {
	color.White("\n════════════════════════════════════════════════════════\n")
	color.Cyan("Optimization Complete!\n")
	color.White("════════════════════════════════════════════════════════\n")

	fmt.Printf("Tasks Completed: %s\n", color.GreenString("%d/%d",
		results.Successful, results.Total))

	if results.Failed > 0 {
		fmt.Printf("Failed: %s\n", color.RedString("%d", results.Failed))
	}

	fmt.Printf("Duration: %s\n", utils.FormatDuration(duration))

	if len(results.CompletedTasks) > 0 {
		color.White("\nCompleted Tasks:\n")
		for _, task := range results.CompletedTasks {
			color.Green("  * %s", task)
		}
	}

	if len(results.FailedTasks) > 0 {
		color.White("\nFailed Tasks:\n")
		for _, task := range results.FailedTasks {
			color.Red("  x %s", task)
		}
	}

	color.White("\n════════════════════════════════════════════════════════\n")
	color.Yellow("\nNote: Some optimizations may require a system restart to take effect.")
}
