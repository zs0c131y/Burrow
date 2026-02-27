package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"github.com/zs0c131y/burrow/pkg/utils"
)

var (
	continuous bool
	interval   int
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Live system monitoring dashboard",
	Long: `Real-time system health monitoring including:
  - CPU usage and per-core breakdown
  - Memory usage and available RAM
  - Disk usage and I/O statistics
  - Network traffic and active connections
  - Top processes by CPU and memory
  - System information and uptime`,
	Run: func(cmd *cobra.Command, args []string) {
		if interval < 1 {
			interval = 2
		}
		runStatus()
	},
}

func init() {
	statusCmd.Flags().BoolVarP(&continuous, "watch", "w", false, "Continuous monitoring mode")
	statusCmd.Flags().IntVarP(&interval, "interval", "i", 2, "Update interval in seconds for watch mode")
}

func runStatus() {
	if continuous {
		runContinuousStatus()
	} else {
		displayStatus()
	}
}

func runContinuousStatus() {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		clearScreen()
		displayStatus()
		fmt.Printf("\n%s Refreshing every %ds (Press Ctrl+C to stop)\n",
			color.YellowString("*"),
			interval)
		<-ticker.C
	}
}

func displayStatus() {
	hostInfo, hostErr := host.Info()
	winVersion := utils.GetWindowsVersion()

	color.Cyan("\n╔════════════════════════════════════════════════════════╗")
	color.Cyan("║                Burrow System Status                    ║")
	color.Cyan("╚════════════════════════════════════════════════════════╝\n")

	healthScore := calculateHealthScore()
	healthColor := getHealthColor(healthScore)

	hostname := "Unknown"
	uptimeStr := "Unknown"
	if hostErr == nil && hostInfo != nil {
		hostname = hostInfo.Hostname
		uptimeStr = utils.FormatDuration(time.Duration(hostInfo.Uptime) * time.Second)
	}

	fmt.Printf("%s %s  %s . %s . Uptime: %s\n\n",
		color.New(color.FgWhite, color.Bold).Sprint("System Health"),
		healthColor.Sprintf("* %d", healthScore),
		hostname,
		winVersion,
		uptimeStr,
	)

	displayCPUInfo()
	fmt.Println()

	displayMemoryInfo()
	fmt.Println()

	displayDiskInfo()
	fmt.Println()

	displayNetworkInfo()
	fmt.Println()

	displayTopProcesses()
}

func displayCPUInfo() {
	cpuPercent, err := cpu.Percent(time.Second, false)
	cpuCounts, _ := cpu.Counts(true)
	cpuInfo, _ := cpu.Info()

	color.New(color.FgCyan, color.Bold).Print("CPU")
	fmt.Println()

	totalPercent := 0.0
	if err == nil && len(cpuPercent) > 0 {
		totalPercent = cpuPercent[0]
	}

	bar := createUsageBar(totalPercent, 20)
	fmt.Printf("   Total   %s  %.1f%%\n", bar, totalPercent)

	if len(cpuInfo) > 0 {
		fmt.Printf("   Model   %s\n", utils.TruncateString(cpuInfo[0].ModelName, 45))
	}
	if cpuCounts > 0 {
		fmt.Printf("   Cores   %d logical processors\n", cpuCounts)
	}

	perCore, err := cpu.Percent(time.Second, true)
	if err == nil && len(perCore) > 0 && len(perCore) <= 64 {
		fmt.Print("   Cores   ")
		for i, p := range perCore {
			barSmall := createUsageBar(p, 8)
			fmt.Printf("C%d %s ", i+1, barSmall)
			if (i+1)%4 == 0 && i != len(perCore)-1 {
				fmt.Print("\n           ")
			}
		}
		fmt.Println()
	}
}

func displayMemoryInfo() {
	memInfo, err := mem.VirtualMemory()
	if err != nil || memInfo == nil {
		color.Red("   Unable to retrieve memory information: %v", err)
		return
	}

	color.New(color.FgCyan, color.Bold).Print("Memory")
	fmt.Println()

	usedPercent := memInfo.UsedPercent
	bar := createUsageBar(usedPercent, 20)

	fmt.Printf("   Used    %s  %.1f%%\n", bar, usedPercent)
	fmt.Printf("   Total   %s\n", utils.FormatBytes(int64(memInfo.Total)))
	fmt.Printf("   Free    %s\n", utils.FormatBytes(int64(memInfo.Available)))
	fmt.Printf("   Cached  %s\n", utils.FormatBytes(int64(memInfo.Cached)))
}

func displayDiskInfo() {
	partitions, err := disk.Partitions(false)
	if err != nil {
		color.Red("   Unable to retrieve disk information: %v", err)
		return
	}

	color.New(color.FgCyan, color.Bold).Print("Disk")
	fmt.Println()

	for _, partition := range partitions {
		if partition.Fstype == "" {
			continue
		}
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil || usage == nil {
			continue
		}

		bar := createUsageBar(usage.UsedPercent, 20)
		fmt.Printf("   %-6s  %s  %.1f%% (%s free)\n",
			partition.Mountpoint,
			bar,
			usage.UsedPercent,
			utils.FormatBytes(int64(usage.Free)),
		)
	}

	ioCounters, err := disk.IOCounters()
	if err == nil && len(ioCounters) > 0 {
		var totalRead, totalWrite uint64
		for _, io := range ioCounters {
			totalRead += io.ReadBytes
			totalWrite += io.WriteBytes
		}
		fmt.Printf("   I/O     Read: %s  Write: %s\n",
			utils.FormatBytes(int64(totalRead)),
			utils.FormatBytes(int64(totalWrite)),
		)
	}
}

func displayNetworkInfo() {
	netIO, err := net.IOCounters(false)
	if err != nil || len(netIO) == 0 {
		color.New(color.FgCyan, color.Bold).Print("Network")
		fmt.Println()
		color.Yellow("   No network data available")
		return
	}

	color.New(color.FgCyan, color.Bold).Print("Network")
	fmt.Println()

	io := netIO[0]
	fmt.Printf("   Sent    %s\n", utils.FormatBytes(int64(io.BytesSent)))
	fmt.Printf("   Recv    %s\n", utils.FormatBytes(int64(io.BytesRecv)))
	fmt.Printf("   Packets Sent: %d  Recv: %d\n", io.PacketsSent, io.PacketsRecv)
}

func displayTopProcesses() {
	color.New(color.FgCyan, color.Bold).Print("Top Processes")
	fmt.Println()

	processes, err := process.Processes()
	if err != nil {
		color.Yellow("   Unable to retrieve process list")
		return
	}

	type procInfo struct {
		Name       string
		PID        int32
		CPUPercent float64
		MemoryMB   uint64
	}

	var procList []procInfo

	for _, p := range processes {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}

		cpuPct, _ := p.CPUPercent()
		memInf, _ := p.MemoryInfo()

		var memMB uint64
		if memInf != nil {
			memMB = memInf.RSS / 1024 / 1024
		}

		procList = append(procList, procInfo{
			Name:       name,
			PID:        p.Pid,
			CPUPercent: cpuPct,
			MemoryMB:   memMB,
		})
	}

	sort.Slice(procList, func(i, j int) bool {
		return procList[i].CPUPercent > procList[j].CPUPercent
	})

	limit := 5
	if len(procList) < limit {
		limit = len(procList)
	}

	for i := 0; i < limit; i++ {
		p := procList[i]
		bar := createUsageBar(p.CPUPercent, 10)
		fmt.Printf("   %-25s %s  CPU: %5.1f%%  RAM: %4dMB\n",
			utils.TruncateString(p.Name, 25),
			bar,
			p.CPUPercent,
			p.MemoryMB,
		)
	}
}

func createUsageBar(percent float64, width int) string {
	if width <= 0 {
		return ""
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("=", filled) + strings.Repeat("-", width-filled)

	if percent < 50 {
		return color.GreenString("[%s]", bar)
	} else if percent < 80 {
		return color.YellowString("[%s]", bar)
	}
	return color.RedString("[%s]", bar)
}

func calculateHealthScore() int {
	score := 100

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		if cpuPercent[0] > 80 {
			score -= 20
		} else if cpuPercent[0] > 60 {
			score -= 10
		}
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil && memInfo != nil {
		if memInfo.UsedPercent > 90 {
			score -= 25
		} else if memInfo.UsedPercent > 75 {
			score -= 15
		}
	}

	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err == nil && usage != nil && usage.UsedPercent > 90 {
				score -= 15
				break
			}
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func getHealthColor(score int) *color.Color {
	if score >= 80 {
		return color.New(color.FgGreen, color.Bold)
	} else if score >= 60 {
		return color.New(color.FgYellow, color.Bold)
	}
	return color.New(color.FgRed, color.Bold)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
