package models

// CleanupTarget represents a path that can be cleaned.
type CleanupTarget struct {
	Name        string
	Path        string
	Description string
	Size        int64
	ItemCount   int
	Category    CleanupCategory
	Protected   bool
}

// CleanupCategory identifies the type of cleanup target.
type CleanupCategory string

const (
	CategoryTemp          CleanupCategory = "Temporary Files"
	CategoryCache         CleanupCategory = "Cache Files"
	CategoryLogs          CleanupCategory = "Log Files"
	CategoryBrowser       CleanupCategory = "Browser Data"
	CategoryWindowsUpdate CleanupCategory = "Windows Update"
	CategoryRecycleBin    CleanupCategory = "Recycle Bin"
	CategoryThumbnails    CleanupCategory = "Thumbnails"
	CategoryPrefetch      CleanupCategory = "Prefetch"
	CategoryDownloads     CleanupCategory = "Downloads"
	CategoryRegistry      CleanupCategory = "Registry"
)

// Application represents an installed Windows application.
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
