package models

import (
	"testing"
)

func TestCleanupCategories(t *testing.T) {
	categories := []CleanupCategory{
		CategoryTemp,
		CategoryCache,
		CategoryLogs,
		CategoryBrowser,
		CategoryWindowsUpdate,
		CategoryRecycleBin,
		CategoryThumbnails,
		CategoryPrefetch,
		CategoryDownloads,
		CategoryRegistry,
	}

	seen := make(map[CleanupCategory]bool)
	for _, c := range categories {
		if c == "" {
			t.Error("Category should not be empty")
		}
		if seen[c] {
			t.Errorf("Duplicate category: %s", c)
		}
		seen[c] = true
	}

	if len(categories) != 10 {
		t.Errorf("Expected 10 categories, got %d", len(categories))
	}
}

func TestCleanupTargetDefaults(t *testing.T) {
	target := CleanupTarget{}
	if target.Size != 0 {
		t.Error("Default Size should be 0")
	}
	if target.Protected {
		t.Error("Default Protected should be false")
	}
	if target.ItemCount != 0 {
		t.Error("Default ItemCount should be 0")
	}
}

func TestApplicationDefaults(t *testing.T) {
	app := Application{}
	if app.Name != "" {
		t.Error("Default Name should be empty")
	}
	if app.Size != 0 {
		t.Error("Default Size should be 0")
	}
}
