package models

import (
	"time"
)

type CleanupTarget struct {
	Name        string
	Path        string
	Description string
	Size        int64
	ItemCount   int
	Category    CleanupCategory
	Protected   bool
}

type CleanupCategory string

const (
	CategoryTemp         CleanupCategory = "Temporary Files"
	CategoryCache        CleanupCategory = "Cache Files"
	CategoryLogs         CleanupCategory = "Log Files"
	CategoryBrowser      CleanupCategory = "Browser Data"
	CategoryWindowsUpdate CleanupCategory = "Windows Update"
	CategoryRecycleBin   CleanupCategory = "Recycle Bin"
	CategoryThumbnails   CleanupCategory = "Thumbnails"
	CategoryPrefetch     CleanupCategory = "Prefetch"
	CategoryDownloads    CleanupCategory = "Downloads"
	CategoryRegistry     CleanupCategory = "Registry"
)

type CleanupResult struct {
	Target        *CleanupTarget
	Success       bool
	FilesRemoved  int
	SpaceFreed    int64
	Error         error
	Duration      time.Duration
}

type CleanupSummary struct {
	TotalTargets    int
	SuccessfulCleans int
	FailedCleans    int
	TotalSpaceFreed int64
	TotalFilesRemoved int
	Duration        time.Duration
	Results         []*CleanupResult
}

type SystemInfo struct {
	OS              string
	Architecture    string
	Hostname        string
	TotalRAM        uint64
	AvailableRAM    uint64
	TotalDisk       uint64
	FreeDisk        uint64
	CPUCount        int
	CPUModel        string
	Uptime          uint64
}

type ProcessInfo struct {
	Name       string
	PID        int32
	CPUPercent float64
	MemoryMB   uint64
}

type DiskUsage struct {
	Path        string
	Size        int64
	ItemCount   int
	IsDirectory bool
	ModTime     time.Time
	Children    []*DiskUsage
}

type Application struct {
	Name            string
	DisplayName     string
	Publisher       string
	Version         string
	InstallDate     string
	InstallLocation string
	UninstallString string
	Size            int64
	RegistryKey     string
}
