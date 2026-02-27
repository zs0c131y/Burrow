package cleanup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zs0c131y/burrow/pkg/models"
)

func TestContains(t *testing.T) {
	slice := []string{"temp", "cache", "logs"}

	tests := []struct {
		item     string
		expected bool
	}{
		{"temp", true},
		{"cache", true},
		{"logs", true},
		{"browser", false},
		{"TEMP", true},
		{"Cache", true},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.item, func(t *testing.T) {
			result := contains(slice, tc.item)
			if result != tc.expected {
				t.Errorf("contains(%v, %q) = %v, want %v", slice, tc.item, result, tc.expected)
			}
		})
	}
}

func TestContainsEmptySlice(t *testing.T) {
	if contains(nil, "anything") {
		t.Error("contains(nil, ...) should return false")
	}
	if contains([]string{}, "anything") {
		t.Error("contains([], ...) should return false")
	}
}

func TestLoadWhitelistEmpty(t *testing.T) {
	os.Setenv("APPDATA", t.TempDir())
	defer os.Unsetenv("APPDATA")

	wl := loadWhitelist()
	if wl == nil {
		t.Fatal("loadWhitelist should return non-nil map")
	}
	if len(wl) != 0 {
		t.Errorf("loadWhitelist should return empty map, got %d entries", len(wl))
	}
}

func TestLoadSaveWhitelist(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "Burrow")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("APPDATA", tmpDir)
	defer os.Unsetenv("APPDATA")

	paths := map[string]bool{
		`c:\important\folder`: true,
		`c:\another\path`:     true,
	}

	if err := saveWhitelist(paths); err != nil {
		t.Fatalf("saveWhitelist error: %v", err)
	}

	wlPath := filepath.Join(configDir, "whitelist.json")
	if _, err := os.Stat(wlPath); os.IsNotExist(err) {
		t.Fatal("whitelist.json not created")
	}

	data, err := os.ReadFile(wlPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg whitelistConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Invalid JSON in whitelist: %v", err)
	}

	if len(cfg.Paths) != 2 {
		t.Errorf("Expected 2 paths in config, got %d", len(cfg.Paths))
	}

	loaded := loadWhitelist()
	if len(loaded) != 2 {
		t.Errorf("loadWhitelist returned %d entries, want 2", len(loaded))
	}

	if !loaded[`c:\important\folder`] {
		t.Error("Expected c:\\important\\folder to be in whitelist")
	}
}

func TestLoadWhitelistCorruptJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "Burrow")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "whitelist.json"), []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("APPDATA", tmpDir)
	defer os.Unsetenv("APPDATA")

	wl := loadWhitelist()
	if len(wl) != 0 {
		t.Errorf("loadWhitelist with corrupt JSON should return empty map, got %d entries", len(wl))
	}
}

func TestIsProtected(t *testing.T) {
	cm := &CleanupManager{
		whitelist: map[string]bool{
			`c:\protected\path`: true,
		},
	}

	if !cm.isProtected(`c:\protected\path`) {
		t.Error("isProtected should return true for whitelisted path")
	}

	if !cm.isProtected(`C:\PROTECTED\PATH`) {
		t.Error("isProtected should be case-insensitive")
	}

	if cm.isProtected(`c:\other\path`) {
		t.Error("isProtected should return false for non-whitelisted path")
	}
}

func TestCleanupManagerInit(t *testing.T) {
	os.Setenv("APPDATA", t.TempDir())
	defer os.Unsetenv("APPDATA")

	cm := NewCleanupManager(true, true)
	if cm == nil {
		t.Fatal("NewCleanupManager returned nil")
	}
	if !cm.debug {
		t.Error("debug should be true")
	}
	if !cm.dryRun {
		t.Error("dryRun should be true")
	}
}

func TestCleanTargetDryRun(t *testing.T) {
	cm := &CleanupManager{
		dryRun:    true,
		whitelist: make(map[string]bool),
	}

	target := &models.CleanupTarget{
		Name:      "Test Target",
		Path:      "/some/path",
		Size:      1000,
		ItemCount: 10,
	}

	result := cm.cleanTarget(target)
	if !result.Success {
		t.Error("Dry run should always succeed")
	}
	if result.SpaceFreed != 1000 {
		t.Errorf("Dry run SpaceFreed = %d, want 1000", result.SpaceFreed)
	}
	if result.FilesRemoved != 10 {
		t.Errorf("Dry run FilesRemoved = %d, want 10", result.FilesRemoved)
	}
}

func TestGetWhitelistPath(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "Burrow")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("APPDATA", tmpDir)
	defer os.Unsetenv("APPDATA")

	path := getWhitelistPath()
	if path == "" {
		t.Error("getWhitelistPath returned empty string")
	}

	if !strings.HasSuffix(path, "whitelist.json") {
		t.Errorf("getWhitelistPath = %q, should end with whitelist.json", path)
	}
}
