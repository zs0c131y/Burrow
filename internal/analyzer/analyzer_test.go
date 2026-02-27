package analyzer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestTree(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	dirs := []string{
		filepath.Join(tmpDir, "dir1"),
		filepath.Join(tmpDir, "dir2"),
		filepath.Join(tmpDir, "dir1", "subdir"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	files := map[string]int{
		filepath.Join(tmpDir, "small.txt"):               100,
		filepath.Join(tmpDir, "dir1", "medium.txt"):      5000,
		filepath.Join(tmpDir, "dir1", "subdir", "a.txt"): 200,
		filepath.Join(tmpDir, "dir2", "large.txt"):       10000,
		filepath.Join(tmpDir, "dir2", "another.txt"):     10000,
	}

	for path, size := range files {
		data := make([]byte, size)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return tmpDir
}

func TestNewAnalyzer(t *testing.T) {
	a := NewAnalyzer(false, false, 3, 0)
	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
	if a.maxDepth != 3 {
		t.Errorf("maxDepth = %d, want 3", a.maxDepth)
	}

	// Test minimum depth clamping
	a2 := NewAnalyzer(false, false, 0, 0)
	if a2.maxDepth != 1 {
		t.Errorf("maxDepth for 0 input = %d, want 1", a2.maxDepth)
	}
}

func TestAnalyzePath(t *testing.T) {
	tmpDir := createTestTree(t)

	a := NewAnalyzer(false, true, 5, 0)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	if tree == nil {
		t.Fatal("AnalyzePath returned nil tree")
	}

	if !tree.IsDirectory {
		t.Error("Root should be directory")
	}

	if tree.ItemCount != 5 {
		t.Errorf("ItemCount = %d, want 5", tree.ItemCount)
	}

	expectedSize := int64(100 + 5000 + 200 + 10000 + 10000)
	if tree.Size != expectedSize {
		t.Errorf("Size = %d, want %d", tree.Size, expectedSize)
	}

	if len(tree.Children) != 3 {
		t.Errorf("Children count = %d, want 3", len(tree.Children))
	}

	// Children should be sorted by size (largest first)
	for i := 0; i < len(tree.Children)-1; i++ {
		if tree.Children[i].Size < tree.Children[i+1].Size {
			t.Errorf("Children not sorted by size: %d < %d",
				tree.Children[i].Size, tree.Children[i+1].Size)
		}
	}
}

func TestAnalyzePathWithMinSize(t *testing.T) {
	tmpDir := createTestTree(t)

	a := NewAnalyzer(false, true, 5, 1000)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	// small.txt (100 bytes) and a.txt (200 bytes) should be excluded
	for _, child := range tree.Children {
		if child.Size < 1000 {
			t.Errorf("Child %s has size %d, should be filtered by minSize 1000",
				child.Name, child.Size)
		}
	}
}

func TestAnalyzePathDepthLimit(t *testing.T) {
	tmpDir := createTestTree(t)

	a := NewAnalyzer(false, true, 1, 0)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	// At depth 1, dir1 and dir2 should not have children expanded
	for _, child := range tree.Children {
		if child.IsDirectory && len(child.Children) > 0 {
			t.Errorf("Depth-1 directory %s should have no children, got %d",
				child.Name, len(child.Children))
		}
	}
}

func TestAnalyzePathNonexistent(t *testing.T) {
	a := NewAnalyzer(false, false, 3, 0)
	_, err := a.AnalyzePath("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestAnalyzePathFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAnalyzer(false, false, 3, 0)
	node, err := a.AnalyzePath(f)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	if node.IsDirectory {
		t.Error("File node should not be directory")
	}
	if node.Size != 5 {
		t.Errorf("Size = %d, want 5", node.Size)
	}
	if node.ItemCount != 1 {
		t.Errorf("ItemCount = %d, want 1", node.ItemCount)
	}
}

func TestGetLargestFiles(t *testing.T) {
	tmpDir := createTestTree(t)

	a := NewAnalyzer(false, true, 5, 0)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	largest := a.GetLargestFiles(tree, 2)
	if len(largest) != 2 {
		t.Fatalf("GetLargestFiles returned %d, want 2", len(largest))
	}

	if largest[0].Size != 10000 {
		t.Errorf("Largest file size = %d, want 10000", largest[0].Size)
	}

	if largest[1].Size != 10000 {
		t.Errorf("Second largest file size = %d, want 10000", largest[1].Size)
	}
}

func TestGetLargestFilesZero(t *testing.T) {
	tmpDir := createTestTree(t)
	a := NewAnalyzer(false, true, 5, 0)
	tree, _ := a.AnalyzePath(tmpDir)

	result := a.GetLargestFiles(tree, 0)
	if result != nil {
		t.Errorf("GetLargestFiles(0) should return nil, got %d items", len(result))
	}
}

func TestGetOldestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	oldFile := filepath.Join(tmpDir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -100)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	newFile := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAnalyzer(false, true, 5, 0)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	oldFiles := a.GetOldestFiles(tree, 30)
	if len(oldFiles) != 1 {
		t.Fatalf("GetOldestFiles returned %d, want 1", len(oldFiles))
	}

	if oldFiles[0].Name != "old.txt" {
		t.Errorf("Oldest file name = %q, want old.txt", oldFiles[0].Name)
	}
}

func TestGetOldestFilesZero(t *testing.T) {
	a := NewAnalyzer(false, false, 5, 0)
	tree := &DiskNode{IsDirectory: true}

	result := a.GetOldestFiles(tree, 0)
	if result != nil {
		t.Errorf("GetOldestFiles(0) should return nil, got %d items", len(result))
	}
}

func TestGetDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	data := []byte("duplicate content here")
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	unique := []byte("unique content that is different")
	if err := os.WriteFile(filepath.Join(tmpDir, "c.txt"), unique, 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAnalyzer(false, true, 5, 0)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	groups := a.GetDuplicates(tree)

	foundDup := false
	for _, g := range groups {
		if len(g.Files) == 2 {
			foundDup = true
			break
		}
	}

	if !foundDup {
		t.Error("GetDuplicates should find the duplicate pair")
	}
}

func TestGetDuplicatesSameSizeDifferentContent(t *testing.T) {
	tmpDir := t.TempDir()

	data1 := []byte("aaaaaaaaaa")
	data2 := []byte("bbbbbbbbbb")
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), data1, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), data2, 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAnalyzer(false, true, 5, 0)
	tree, err := a.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath error: %v", err)
	}

	groups := a.GetDuplicates(tree)

	for _, g := range groups {
		if len(g.Files) >= 2 {
			t.Error("Files with same size but different content should not be grouped as duplicates")
		}
	}
}

func TestHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hide"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "visible"), []byte("show"), 0o644); err != nil {
		t.Fatal(err)
	}

	aNoHidden := NewAnalyzer(false, false, 5, 0)
	tree1, _ := aNoHidden.AnalyzePath(tmpDir)
	if tree1.ItemCount != 1 {
		t.Errorf("Without hidden: ItemCount = %d, want 1", tree1.ItemCount)
	}

	aWithHidden := NewAnalyzer(false, true, 5, 0)
	tree2, _ := aWithHidden.AnalyzePath(tmpDir)
	if tree2.ItemCount != 2 {
		t.Errorf("With hidden: ItemCount = %d, want 2", tree2.ItemCount)
	}
}
