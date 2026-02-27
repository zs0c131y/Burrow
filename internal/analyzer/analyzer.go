package analyzer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zs0c131y/burrow/pkg/utils"
)

// Analyzer performs disk space analysis.
type Analyzer struct {
	debug      bool
	showHidden bool
	maxDepth   int
	minSize    int64
}

// DiskNode represents a file or directory in the analysis tree.
type DiskNode struct {
	Name        string
	Path        string
	Size        int64
	ItemCount   int
	IsDirectory bool
	Children    []*DiskNode
	LargeFiles  int
	ModTime     time.Time
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer(debug, showHidden bool, maxDepth int, minSize int64) *Analyzer {
	if maxDepth < 1 {
		maxDepth = 1
	}
	return &Analyzer{
		debug:      debug,
		showHidden: showHidden,
		maxDepth:   maxDepth,
		minSize:    minSize,
	}
}

// AnalyzePath analyzes the given path and returns a tree of DiskNodes.
func (a *Analyzer) AnalyzePath(path string) (*DiskNode, error) {
	return a.analyzeNode(path, 0)
}

func (a *Analyzer) analyzeNode(path string, depth int) (*DiskNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot stat %s: %w", path, err)
	}

	node := &DiskNode{
		Name:        filepath.Base(path),
		Path:        path,
		IsDirectory: info.IsDir(),
		ModTime:     info.ModTime(),
	}

	if !info.IsDir() {
		node.Size = info.Size()
		node.ItemCount = 1

		if node.Size > 100*1024*1024 {
			node.LargeFiles = 1
		}

		return node, nil
	}

	if depth >= a.maxDepth {
		size, count, err := utils.GetDirSize(path)
		if err != nil {
			return nil, fmt.Errorf("cannot calculate size of %s: %w", path, err)
		}
		node.Size = size
		node.ItemCount = count
		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		size, count, _ := utils.GetDirSize(path)
		node.Size = size
		node.ItemCount = count
		return node, nil
	}

	for _, entry := range entries {
		if !a.showHidden && isHidden(entry.Name()) {
			continue
		}

		childPath := filepath.Join(path, entry.Name())

		childNode, err := a.analyzeNode(childPath, depth+1)
		if err != nil {
			continue
		}

		if childNode.Size >= a.minSize {
			node.Children = append(node.Children, childNode)
			node.Size += childNode.Size
			node.ItemCount += childNode.ItemCount
			node.LargeFiles += childNode.LargeFiles
		}
	}

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

// GetLargestFiles returns the N largest files from the tree.
func (a *Analyzer) GetLargestFiles(root *DiskNode, n int) []*DiskNode {
	if n <= 0 {
		return nil
	}

	var files []*DiskNode
	a.collectFiles(root, &files)

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

// GetOldestFiles returns files not modified in the last daysOld days.
func (a *Analyzer) GetOldestFiles(root *DiskNode, daysOld int) []*DiskNode {
	if daysOld <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -daysOld)
	var oldFiles []*DiskNode

	var walk func(node *DiskNode)
	walk = func(node *DiskNode) {
		if !node.IsDirectory {
			if !node.ModTime.IsZero() && node.ModTime.Before(cutoff) {
				oldFiles = append(oldFiles, node)
			}
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(root)

	sort.Slice(oldFiles, func(i, j int) bool {
		return oldFiles[i].ModTime.Before(oldFiles[j].ModTime)
	})

	return oldFiles
}

// DuplicateGroup represents a set of files that are potential duplicates.
type DuplicateGroup struct {
	Size  int64
	Hash  string
	Files []*DiskNode
}

// GetDuplicates finds duplicate files by first matching on size, then verifying
// with a partial content hash (first 8KB) for efficiency.
func (a *Analyzer) GetDuplicates(root *DiskNode) []DuplicateGroup {
	sizeMap := make(map[int64][]*DiskNode)

	var collectBySize func(*DiskNode)
	collectBySize = func(node *DiskNode) {
		if !node.IsDirectory && node.Size > 0 {
			sizeMap[node.Size] = append(sizeMap[node.Size], node)
		}
		for _, child := range node.Children {
			collectBySize(child)
		}
	}

	collectBySize(root)

	var groups []DuplicateGroup

	for size, nodes := range sizeMap {
		if len(nodes) < 2 {
			continue
		}

		hashGroups := make(map[string][]*DiskNode)
		for _, node := range nodes {
			hash, err := partialHash(node.Path)
			if err != nil {
				continue
			}
			hashGroups[hash] = append(hashGroups[hash], node)
		}

		for hash, hNodes := range hashGroups {
			if len(hNodes) >= 2 {
				groups = append(groups, DuplicateGroup{
					Size:  size,
					Hash:  hash,
					Files: hNodes,
				})
			}
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Size*int64(len(groups[i].Files)) > groups[j].Size*int64(len(groups[j].Files))
	})

	return groups
}

// partialHash computes a SHA-256 hash of the first 8KB of a file for quick comparison.
func partialHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	h := sha256.Sum256(buf[:n])
	return fmt.Sprintf("%x", h), nil
}

// isHidden checks if a file/directory name indicates it's hidden.
// On Windows, files starting with '.' or having the hidden attribute are hidden.
func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}
