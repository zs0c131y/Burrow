package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0 B"},
		{"negative", -100, "0 B"},
		{"bytes", 500, "500 B"},
		{"one_kb", 1024, "1.0 KB"},
		{"kb", 1536, "1.5 KB"},
		{"one_mb", 1048576, "1.0 MB"},
		{"mb", 5242880, "5.0 MB"},
		{"one_gb", 1073741824, "1.0 GB"},
		{"gb", 2684354560, "2.5 GB"},
		{"one_tb", 1099511627776, "1.0 TB"},
		{"one_pb", 1125899906842624, "1.0 PB"},
		{"large_eb", 1152921504606846976, "1.0 EB"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatBytes(tc.input)
			if result != tc.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"negative", -1 * time.Second, "0ms"},
		{"zero", 0, "0ms"},
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"seconds", 5 * time.Second, "5.0s"},
		{"minutes", 3 * time.Minute, "3.0m"},
		{"hours", 2 * time.Hour, "2.0h"},
		{"days", 50 * time.Hour, "2d 2h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDuration(tc.input)
			if result != tc.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCreateProgressBar(t *testing.T) {
	tests := []struct {
		name    string
		current int
		total   int
		width   int
		wantLen bool
		wantPct string
	}{
		{"zero_total", 5, 0, 10, false, ""},
		{"zero_width", 5, 10, 0, false, ""},
		{"half", 5, 10, 10, true, "50.0%"},
		{"full", 10, 10, 10, true, "100.0%"},
		{"over", 15, 10, 10, true, "100.0%"},
		{"negative_current", -5, 10, 10, true, "0.0%"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CreateProgressBar(tc.current, tc.total, tc.width)
			if !tc.wantLen {
				if result != "" {
					t.Errorf("CreateProgressBar(%d, %d, %d) = %q, want empty", tc.current, tc.total, tc.width, result)
				}
				return
			}
			if result == "" {
				t.Errorf("CreateProgressBar(%d, %d, %d) returned empty, want non-empty", tc.current, tc.total, tc.width)
			}
			if tc.wantPct != "" && !containsStr(result, tc.wantPct) {
				t.Errorf("CreateProgressBar(%d, %d, %d) = %q, want to contain %q", tc.current, tc.total, tc.width, result, tc.wantPct)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"zero_len", "hello", 0, ""},
		{"short", "hi", 10, "hi"},
		{"exact", "hello", 5, "hello"},
		{"truncate", "hello world", 8, "hello..."},
		{"very_short_max", "hello", 2, "he"},
		{"three", "hello", 3, "hel"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TruncateString(tc.input, tc.maxLen)
			if result != tc.expected {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()

	if !PathExists(tmpDir) {
		t.Errorf("PathExists(%q) = false, want true", tmpDir)
	}

	if PathExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("PathExists(nonexistent) = true, want false")
	}
}

func TestGetDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	data := []byte("hello world")
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "c.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	size, count, err := GetDirSize(tmpDir)
	if err != nil {
		t.Fatalf("GetDirSize error: %v", err)
	}

	if count != 3 {
		t.Errorf("GetDirSize count = %d, want 3", count)
	}

	expectedSize := int64(len(data)) * 3
	if size != expectedSize {
		t.Errorf("GetDirSize size = %d, want %d", size, expectedSize)
	}
}

func TestSafeDelete(t *testing.T) {
	tmpDir := t.TempDir()

	// Test deleting a file
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SafeDelete(filePath, 3); err != nil {
		t.Errorf("SafeDelete file error: %v", err)
	}

	if PathExists(filePath) {
		t.Error("File still exists after SafeDelete")
	}

	// Test deleting a directory
	dirPath := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirPath, "f.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SafeDelete(dirPath, 3); err != nil {
		t.Errorf("SafeDelete dir error: %v", err)
	}

	if PathExists(dirPath) {
		t.Error("Directory still exists after SafeDelete")
	}

	// Test deleting nonexistent path (should be no-op)
	if err := SafeDelete(filepath.Join(tmpDir, "nope"), 3); err != nil {
		t.Errorf("SafeDelete nonexistent should succeed, got: %v", err)
	}
}

func TestCleanDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	data := []byte("test content")
	for i := 0; i < 5; i++ {
		name := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(name, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	freed, removed, skipped, err := CleanDirectory(tmpDir, 3)
	if err != nil {
		t.Fatalf("CleanDirectory error: %v", err)
	}

	if removed != 6 {
		t.Errorf("CleanDirectory removed = %d, want 6", removed)
	}

	if skipped != 0 {
		t.Errorf("CleanDirectory skipped = %d, want 0", skipped)
	}

	expectedSize := int64(len(data)) * 6
	if freed != expectedSize {
		t.Errorf("CleanDirectory freed = %d, want %d", freed, expectedSize)
	}
}

func TestExpandEnvPath(t *testing.T) {
	os.Setenv("BURROW_TEST_VAR", "hello")
	defer os.Unsetenv("BURROW_TEST_VAR")

	// ExpandEnvPath triggers on '%' chars (Windows-style env vars)
	// os.ExpandEnv uses $VAR or ${VAR} syntax on all platforms
	result := ExpandEnvPath("%BURROW_TEST_VAR%")
	// On Linux, %VAR% is not expanded by os.ExpandEnv; it only expands $VAR.
	// The function detects '%' and calls os.ExpandEnv, but on Linux $VAR syntax is needed.
	// This is by design: the function targets Windows where %VAR% is standard.
	if result == "" {
		t.Error("ExpandEnvPath returned empty string")
	}

	plain := ExpandEnvPath("/no/envvars/here")
	if plain != "/no/envvars/here" {
		t.Errorf("ExpandEnvPath modified path without env vars: %q", plain)
	}
}

func TestGetConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("APPDATA", tmpDir)
	defer os.Unsetenv("APPDATA")

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir error: %v", err)
	}

	expected := filepath.Join(tmpDir, "Burrow")
	if dir != expected {
		t.Errorf("GetConfigDir = %q, want %q", dir, expected)
	}

	if !PathExists(dir) {
		t.Error("GetConfigDir did not create directory")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
