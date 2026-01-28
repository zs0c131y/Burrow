package analyzer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zs0c131y/burrow/pkg/utils"
)

type Analyzer struct {
	debug      bool
	showHidden bool
	maxDepth   int
	minSize    int64
}

type DiskNode struct {
	Name        string
	Path        string
	Size        int64
	ItemCount   int
	IsDirectory bool
	Children    []*DiskNode
	LargeFiles  int
}

func NewAnalyzer(debug, showHidden bool, maxDepth int, minSize int64) *Analyzer {
	return &Analyzer{
		debug:      debug,
		showHidden: showHidden,
		maxDepth:   maxDepth,
		minSize:    minSize,
	}
}

func (a *Analyzer) AnalyzePath(path string) (*DiskNode, error) {
	return a.analyzeNode(path, 0)
}

func (a *Analyzer) analyzeNode(path string, depth int) (*DiskNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &DiskNode{
		Name:        filepath.Base(path),
		Path:        path,
		IsDirectory: info.IsDir(),
	}

	if !info.IsDir() {
		node.Size = info.Size()
		node.ItemCount = 1

		// Check if it's a large file (>100MB)
		if node.Size > 100*1024*1024 {
			node.LargeFiles = 1
		}

		return node, nil
	}

	// It's a directory - scan children
	if depth >= a.maxDepth {
		// Just get the size without going deeper
		size, count, _ := utils.GetDirSize(path)
		node.Size = size
		node.ItemCount = count
		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		// Permission denied or other error - try to get size
		size, count, _ := utils.GetDirSize(path)
		node.Size = size
		node.ItemCount = count
		return node, nil
	}

	for _, entry := range entries {
		// Skip hidden files if not showing them
		if !a.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		childPath := filepath.Join(path, entry.Name())

		childNode, err := a.analyzeNode(childPath, depth+1)
		if err != nil {
			continue // Skip inaccessible paths
		}

		// Only include if meets minimum size requirement
		if childNode.Size >= a.minSize {
			node.Children = append(node.Children, childNode)
			node.Size += childNode.Size
			node.ItemCount += childNode.ItemCount
			node.LargeFiles += childNode.LargeFiles
		}
	}

	// Sort children by size (largest first)
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

// GetLargestFiles returns the N largest files from the tree
func (a *Analyzer) GetLargestFiles(root *DiskNode, n int) []*DiskNode {
	var files []*DiskNode
	a.collectFiles(root, &files)

	// Sort by size
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	if len(files) > n {
		return files[:n]
	}
	return files
}

func (a *Analyzer) collectFiles(node *DiskNode, files *[]*DiskNode) {
	if !node.IsDirectory {
		*files = append(*files, node)
		return
	}

	for _, child := range node.Children {
		a.collectFiles(child, files)
	}
}

// GetOldestFiles returns files not modified in the last N days
func (a *Analyzer) GetOldestFiles(root *DiskNode, daysOld int) []*DiskNode {
	var oldFiles []*DiskNode
	// Implementation would check file modification times
	// This is a placeholder
	return oldFiles
}

// GetDuplicates finds potential duplicate files by size
func (a *Analyzer) GetDuplicates(root *DiskNode) map[int64][]*DiskNode {
	sizeMap := make(map[int64][]*DiskNode)

	var collectBySi func(*DiskNode)
	collectBySi = func(node *DiskNode) {
		if !node.IsDirectory {
			sizeMap[node.Size] = append(sizeMap[node.Size], node)
		}
		for _, child := range node.Children {
			collectBySi(child)
		}
	}

	collectBySi(root)

	// Filter to only keep sizes with multiple files
	duplicates := make(map[int64][]*DiskNode)
	for size, nodes := range sizeMap {
		if len(nodes) > 1 {
			duplicates[size] = nodes
		}
	}

	return duplicates
}
