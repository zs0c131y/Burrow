package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/zs0c131y/burrow/internal/uninstall"
	"github.com/zs0c131y/burrow/pkg/utils"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Smart app removal with leftover cleanup",
	Long: `Intelligently uninstall applications including:
  • Application files and executables
  • Registry entries and keys
  • AppData and LocalAppData remnants
  • Start menu and desktop shortcuts
  • Service entries
  • Scheduled tasks
  • Temp files and caches`,
	Run: func(cmd *cobra.Command, args []string) {
		runUninstall()
	},
}

func runUninstall() {
	utils.RequireAdmin()

	color.Cyan("\n╔════════════════════════════════════════════════════════╗")
	color.Cyan("║             Burrow Smart App Uninstaller               ║")
	color.Cyan("╚════════════════════════════════════════════════════════╝\n")

	color.White("Discovering installed applications...\n")

	manager := uninstall.NewUninstallManager(debugMode, dryRun)

	apps, err := manager.DiscoverApplications()
	if err != nil {
		color.Red("Error discovering applications: %v", err)
		return
	}

	if len(apps) == 0 {
		color.Yellow("No applications found.")
		return
	}

	color.Green("Found %d installed applications\n", len(apps))

	// Create selection menu
	var menuItems []string
	for _, app := range apps {
		sizeStr := "Unknown size"
		if app.Size > 0 {
			sizeStr = utils.FormatBytes(app.Size)
		}
		menuItems = append(menuItems, fmt.Sprintf("%s (%s) - %s",
			app.DisplayName,
			sizeStr,
			app.Publisher,
		))
	}

	prompt := promptui.Select{
		Label: "Select application to uninstall",
		Items: menuItems,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		color.Yellow("Selection cancelled")
		return
	}

	selectedApp := apps[index]

	fmt.Println()
	color.White("Selected: %s\n", color.New(color.FgCyan, color.Bold).Sprint(selectedApp.DisplayName))
	color.White("Publisher: %s", selectedApp.Publisher)
	color.White("Version: %s", selectedApp.Version)
	if selectedApp.Size > 0 {
		color.White("Size: %s", utils.FormatBytes(selectedApp.Size))
	}
	fmt.Println()

	if dryRun {
		color.Yellow("DRY RUN: Showing what would be removed\n")
		manager.PreviewUninstall(selectedApp)
		return
	}

	if !confirmAction(fmt.Sprintf("Uninstall %s and remove all leftovers?", selectedApp.DisplayName)) {
		color.Yellow("Uninstall cancelled")
		return
	}

	fmt.Println()
	color.White("Uninstalling %s...\n", selectedApp.DisplayName)

	result := manager.UninstallApplication(selectedApp)

	displayUninstallResult(result)
}

func displayUninstallResult(result *uninstall.UninstallResult) {
	color.White("\n════════════════════════════════════════════════════════\n")

	if result.Success {
		color.Green("✓ Uninstall Complete!\n")
	} else {
		color.Red("✗ Uninstall Failed\n")
		if result.Error != nil {
			color.Red("Error: %v\n", result.Error)
		}
		color.White("════════════════════════════════════════════════════════\n")
		return
	}

	color.White("════════════════════════════════════════════════════════\n")

	fmt.Printf("Files Removed: %d\n", result.FilesRemoved)
	fmt.Printf("Registry Keys Removed: %d\n", result.RegistryKeysRemoved)
	fmt.Printf("Space Freed: %s\n", color.GreenString(utils.FormatBytes(result.SpaceFreed)))
	fmt.Printf("Duration: %s\n", utils.FormatDuration(result.Duration))

	if len(result.LocationsCleaned) > 0 {
		color.White("\nLocations Cleaned:\n")
		for _, loc := range result.LocationsCleaned {
			color.Green("  ✓ %s", loc)
		}
	}

	color.White("\n════════════════════════════════════════════════════════\n")
}
