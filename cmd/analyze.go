package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zs0c131y/burrow/internal/analyzer"
	"github.com/zs0c131y/burrow/pkg/utils"
)

var (
	analyzePath  string
	analyzeDepth int
	showHidden   bool
	minSize      int64
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Visual disk space analyzer",
	Long: `Analyze disk usage with visual tree display:
  - Interactive directory explorer
  - Size-based sorting and filtering
  - Large file identification
  - Visual percentage bars`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			analyzePath = args[0]
		}
		runAnalyze()
	},
}

func init() {
	analyzeCmd.Flags().IntVarP(&analyzeDepth, "depth", "d", 3, "Maximum depth to analyze (1-10)")
	analyzeCmd.Flags().BoolVar(&showHidden, "hidden", false, "Show hidden files and folders")
	analyzeCmd.Flags().Int64Var(&minSize, "min-size", 0, "Minimum size in MB to display")
}

func runAnalyze() {
	if analyzePath == "" {
		analyzePath = "C:\\"
	}

	if analyzeDepth < 1 {
		analyzeDepth = 1
	}
	if analyzeDepth > 10 {
		analyzeDepth = 10
	}

	if minSize < 0 {
		minSize = 0
	}

	absPath, err := filepath.Abs(analyzePath)
	if err != nil {
		color.Red("Invalid path: %v", err)
		return
	}

	if !utils.PathExists(absPath) {
		color.Red("Path does not exist: %s", absPath)
		return
	}

	color.Cyan("\n╔════════════════════════════════════════════════════════╗")
	color.Cyan("║              Burrow Disk Space Analyzer                ║")
	color.Cyan("╚════════════════════════════════════════════════════════╝\n")

	color.White("Analyzing: %s\n", color.CyanString(absPath))
	color.White("Depth: %d | Min size: %s | Hidden: %v\n",
		analyzeDepth,
		utils.FormatBytes(minSize*1024*1024),
		showHidden,
	)
	color.White("Please wait, scanning directory tree...\n\n")

	a := analyzer.NewAnalyzer(debugMode, showHidden, analyzeDepth, minSize*1024*1024)

	tree, err := a.AnalyzePath(absPath)
	if err != nil {
		color.Red("Error analyzing path: %v", err)
		return
	}

	displayAnalysis(tree, absPath)

	largeFiles := a.GetLargestFiles(tree, 10)
	if len(largeFiles) > 0 {
		displayLargeFiles(largeFiles)
	}
}

func displayAnalysis(tree *analyzer.DiskNode, rootPath string) {
	color.White("════════════════════════════════════════════════════════\n")
	fmt.Printf("Path: %s\n", color.CyanString(rootPath))
	fmt.Printf("Total Size: %s\n", color.New(color.FgGreen, color.Bold).Sprint(utils.FormatBytes(tree.Size)))
	fmt.Printf("Items: %s files and folders\n", color.WhiteString("%d", tree.ItemCount))
	color.White("════════════════════════════════════════════════════════\n\n")

	displayTopItems(tree, 20)
}

func displayTopItems(node *analyzer.DiskNode, limit int) {
	if len(node.Children) == 0 {
		color.Yellow("No items to display")
		return
	}

	totalSize := node.Size
	if totalSize <= 0 {
		totalSize = 1
	}

	color.New(color.FgCyan, color.Bold).Println("Top Space Consumers:")
	fmt.Println()

	displayCount := limit
	if len(node.Children) < limit {
		displayCount = len(node.Children)
	}

	for i := 0; i < displayCount; i++ {
		child := node.Children[i]

		percentage := float64(child.Size) / float64(totalSize) * 100
		if percentage > 100 {
			percentage = 100
		}

		barWidth := 20
		filled := int(percentage / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		if filled < 0 {
			filled = 0
		}

		bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)

		icon := "F"
		if child.IsDirectory {
			icon = "D"
		}

		barColor := color.GreenString
		if percentage > 20 {
			barColor = color.YellowString
		}
		if percentage > 40 {
			barColor = color.RedString
		}

		fmt.Printf(" %2d. %s  %5.1f%%  %s  %-40s %10s",
			i+1,
			barColor("["+bar+"]"),
			percentage,
			icon,
			utils.TruncateString(child.Name, 40),
			color.CyanString(utils.FormatBytes(child.Size)),
		)

		if child.IsDirectory && child.ItemCount > 0 {
			fmt.Printf("  (%d items)", child.ItemCount)
		}

		fmt.Println()
	}

	fmt.Println()
	color.White("════════════════════════════════════════════════════════\n")

	if node.LargeFiles > 0 {
		color.Yellow("Found %d files larger than 100MB", node.LargeFiles)
	}
}

func displayLargeFiles(files []*analyzer.DiskNode) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("Largest Files:")
	fmt.Println()

	for i, f := range files {
		fmt.Printf("  %2d. %-50s %10s\n",
			i+1,
			utils.TruncateString(f.Path, 50),
			color.CyanString(utils.FormatBytes(f.Size)),
		)
	}

	color.White("\n════════════════════════════════════════════════════════\n")
}
