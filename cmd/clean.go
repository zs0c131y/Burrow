package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zs0c131y/burrow/internal/cleanup"
	"github.com/zs0c131y/burrow/pkg/utils"
)

var (
	whitelistMode bool
	categories    []string
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Deep cleanup of temporary files, caches, and logs",
	Long: `Performs comprehensive system cleanup including:
  - Temporary files (Windows Temp, User Temp)
  - Browser caches (Chrome, Firefox, Edge, Brave)
  - Windows Update cache
  - Application caches
  - System logs and event logs
  - Recycle Bin
  - Thumbnails and icon cache
  - Prefetch files`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanup()
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&whitelistMode, "whitelist", false, "Manage protected paths that won't be cleaned")
	cleanCmd.Flags().StringSliceVar(&categories, "categories", []string{}, "Specific categories to clean (temp,cache,logs,browser,updates)")
}

func runCleanup() {
	if whitelistMode {
		cleanup.ManageWhitelist()
		return
	}

	if !dryRun {
		if err := utils.RequireAdmin(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	}

	color.Cyan("\n╔════════════════════════════════════════════════════════╗")
	color.Cyan("║             Burrow Deep System Cleanup                 ║")
	color.Cyan("╚════════════════════════════════════════════════════════╝\n")

	if dryRun {
		color.Yellow("DRY RUN MODE - No files will be deleted\n")
	}

	color.White("Scanning system for cleanup targets...\n")

	startTime := time.Now()

	manager := cleanup.NewCleanupManager(debugMode, dryRun)

	targets, err := manager.DiscoverTargets(categories)
	if err != nil {
		color.Red("Error discovering cleanup targets: %v", err)
		return
	}

	if len(targets) == 0 {
		color.Green("System is already clean!")
		return
	}

	color.White("\nDiscovered Cleanup Targets:\n")
	color.White("════════════════════════════════════════════════════════\n")

	totalSize := int64(0)
	totalFiles := 0

	for _, target := range targets {
		statusIcon := "*"
		statusColor := color.GreenString
		if target.Protected {
			statusIcon = "!"
			statusColor = color.YellowString
		}

		fmt.Printf("  %s %-40s %10s (%d files)\n",
			statusColor(statusIcon),
			utils.TruncateString(target.Name, 40),
			color.CyanString(utils.FormatBytes(target.Size)),
			target.ItemCount,
		)

		if !target.Protected {
			totalSize += target.Size
			totalFiles += target.ItemCount
		}
	}

	color.White("\n════════════════════════════════════════════════════════\n")
	fmt.Printf("Total Space to Free: %s | Files: %d\n",
		color.New(color.FgGreen, color.Bold).Sprint(utils.FormatBytes(totalSize)),
		totalFiles,
	)
	color.White("════════════════════════════════════════════════════════\n\n")

	if dryRun {
		color.Yellow("Dry run complete. No changes were made.")
		return
	}

	if !confirmAction("Proceed with cleanup?") {
		color.Yellow("Cleanup cancelled.")
		return
	}

	fmt.Println()
	color.White("Cleaning system...\n")

	summary := manager.ExecuteCleanup(targets)

	displayCleanupResults(summary, time.Since(startTime))
}

func displayCleanupResults(summary *cleanup.CleanupSummary, duration time.Duration) {
	color.White("\n════════════════════════════════════════════════════════\n")
	color.Cyan("Cleanup Complete!\n")
	color.White("════════════════════════════════════════════════════════\n")

	fmt.Printf("Targets Processed: %d\n", summary.TotalTargets)
	fmt.Printf("Successful: %s\n", color.GreenString("%d", summary.SuccessfulCleans))

	if summary.FailedCleans > 0 {
		fmt.Printf("Failed: %s\n", color.RedString("%d", summary.FailedCleans))
	}

	fmt.Printf("\n%s %s\n",
		color.New(color.FgGreen, color.Bold).Sprint("Space Freed:"),
		color.New(color.FgGreen, color.Bold).Sprint(utils.FormatBytes(summary.TotalSpaceFreed)),
	)
	fmt.Printf("Files Removed: %s\n", color.CyanString("%d", summary.TotalFilesRemoved))
	fmt.Printf("Duration: %s\n", color.WhiteString(utils.FormatDuration(duration)))

	if debugMode && len(summary.Results) > 0 {
		color.White("\nDetailed Results:\n")
		for _, result := range summary.Results {
			if !result.Success {
				color.Red("  x %s: %v", result.Target.Name, result.Error)
			} else {
				color.Green("  * %s: %s", result.Target.Name, utils.FormatBytes(result.SpaceFreed))
			}
		}
	}

	color.White("════════════════════════════════════════════════════════\n")
}

func confirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	r := trimLower(response)
	return r == "y" || r == "yes"
}

func trimLower(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result = append(result, c)
		}
	}
	return string(result)
}
