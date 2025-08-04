// internal/layout/layout.go
package layout

import (
	"hash/fnv"
	"sort"
	"strings"

	"example.com/village-watch/internal/domain"
)

type Slot struct {
	X, Y int
	W, H int
	Path string
}

// VillageLayout creates a village-like arrangement with districts and streets
func Grid(root *domain.FileNode, cols, rows int) []Slot {
	items := collect(root)
	
	// Separate directories (districts) from files
	var districts []*domain.FileNode
	var files []*domain.FileNode
	
	for _, item := range items {
		if item.IsDir {
			districts = append(districts, item)
		} else {
			files = append(files, item)
		}
	}
	
	// Sort both groups
	sort.Slice(districts, func(i, j int) bool { return strings.ToLower(districts[i].Name) < strings.ToLower(districts[j].Name) })
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name) })
	
	slots := []Slot{}
	
	// Layout districts first with spacing
	x, y := 0, 0
	
	for _, district := range districts {
		// Districts get 3x2 size with spacing
		w, h := min(3, cols-x), 2
		if x+w > cols {
			x = 0
			y += 3 // Leave space between district rows
			if y >= rows {
				break
			}
			w = min(3, cols)
		}
		
		slots = append(slots, Slot{X: x, Y: y, W: w, H: h, Path: district.Path})
		x += w + 1 // Add spacing between districts
	}
	
	// Start files on a new "street" below districts
	if len(districts) > 0 {
		y += 3 // Street spacing after districts
		x = 0
	}
	
	// Layout files in streets
	for _, file := range files {
		if x >= cols {
			x = 0
			y += 1
			if y >= rows {
				break
			}
		}
		
		slots = append(slots, Slot{X: x, Y: y, W: 1, H: 1, Path: file.Path})
		x += 1
	}
	
	return slots
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func collect(n *domain.FileNode) []*domain.FileNode {
	var out []*domain.FileNode
	var walk func(*domain.FileNode)
	walk = func(cur *domain.FileNode) {
		if cur == nil {
			return
		}
		for _, ch := range cur.Children {
			out = append(out, ch)
			if ch.IsDir {
				walk(ch)
			}
		}
	}
	walk(n)
	return out
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
