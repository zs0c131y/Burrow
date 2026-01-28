package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	debugMode  bool
	dryRun     bool
	appVersion string
)

var rootCmd = &cobra.Command{
	Use:   "wm",
	Short: "Burrow - Deep Windows System Optimizer",
	Long: color.CyanString(`
╔╗ ╦ ╦╦═╗╦═╗╔═╗╦ ╦
╠╩╗║ ║╠╦╝╠╦╝║ ║║║║
╚═╝╚═╝╩╚═╩╚═╚═╝╚╩╝

Dig deep like a mole to optimize your Windows system.
All-in-one: Cleaner + Uninstaller + Monitor + Optimizer + Analyzer
`),
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "windows" {
			color.Red("Burrow is designed for Windows systems only.")
			os.Exit(1)
		}
		showInteractiveMenu()
	},
}

func Execute(version string) error {
	appVersion = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode with detailed logs")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview changes without making them")

	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(optimizeCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}

func showInteractiveMenu() {
	banner := color.New(color.FgCyan, color.Bold)
	banner.Print("\n╔════════════════════════════════════════════════════════╗\n")
	banner.Print("║             Burrow - Windows System Optimizer          ║\n")
	banner.Print("╚════════════════════════════════════════════════════════╝\n\n")

	menuItems := []string{
		"🧹 Deep Cleanup - Remove temp files, caches, logs",
		"🗑️  Uninstall Apps - Smart removal with leftover cleanup",
		"⚡ Optimize System - Services, startup, registry tuning",
		"📊 System Status - Live CPU, RAM, disk, network monitor",
		"💾 Disk Analyzer - Visual space usage explorer",
		"ℹ️  About & Version",
		"❌ Exit",
	}

	prompt := promptui.Select{
		Label:        "Select an action",
		Items:        menuItems,
		Size:         7,
		HideSelected: false,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	for {
		index, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Menu selection error: %v\n", err)
			return
		}

		fmt.Println()

		switch index {
		case 0:
			cleanCmd.Run(cleanCmd, []string{})
		case 1:
			uninstallCmd.Run(uninstallCmd, []string{})
		case 2:
			optimizeCmd.Run(optimizeCmd, []string{})
		case 3:
			statusCmd.Run(statusCmd, []string{})
		case 4:
			analyzeCmd.Run(analyzeCmd, []string{})
		case 5:
			versionCmd.Run(versionCmd, []string{})
		case 6:
			color.Yellow("\nThank you for using Burrow!\n")
			os.Exit(0)
		}

		fmt.Println()
		color.Cyan("Press Enter to return to menu...")
		fmt.Scanln()
		fmt.Println()
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Burrow version",
	Run: func(cmd *cobra.Command, args []string) {
		color.Cyan("Burrow v%s", appVersion)
		color.White("Windows System Optimizer")
		color.White("Built with Go %s", runtime.Version())
		color.White("OS: %s/%s", runtime.GOOS, runtime.GOARCH)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Burrow to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		color.Yellow("Checking for updates...")
		color.White("Current version: %s", appVersion)
		color.Green("\nTo update manually, download the latest release from:")
		color.Cyan("https://github.com/zs0c131y/burrow/releases")
	},
}
