// internal/layout/layout.go
package layout

import (
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"

	"example.com/village-watch/internal/domain"
)

type Slot struct {
	X, Y int
	W, H int
	Path string
}

// VillageLayout creates a village-like arrangement with max 10 buildings, prioritizing top-level folders
func Grid(root *domain.FileNode, cols, rows int) []Slot {
	if root == nil {
		return []Slot{}
	}
	
	// Get only direct children (top-level items)
	topLevel := root.Children
	if len(topLevel) == 0 {
		return []Slot{}
	}
	
	// Separate directories from files at top level
	var topDirs []*domain.FileNode
	var topFiles []*domain.FileNode
	
	for _, item := range topLevel {
		if item.IsDir {
			topDirs = append(topDirs, item)
		} else {
			topFiles = append(topFiles, item)
		}
	}
	
	// Sort by name
	sort.Slice(topDirs, func(i, j int) bool { return strings.ToLower(topDirs[i].Name) < strings.ToLower(topDirs[j].Name) })
	sort.Slice(topFiles, func(i, j int) bool { return strings.ToLower(topFiles[i].Name) < strings.ToLower(topFiles[j].Name) })
	
	// Select up to 10 buildings total, prioritizing directories
	var selected []*domain.FileNode
	maxBuildings := 10
	
	// Add all top-level directories first
	for _, dir := range topDirs {
		if len(selected) >= maxBuildings {
			break
		}
		selected = append(selected, dir)
	}
	
	// If we have room, add some subdirectories from the largest top-level dirs
	remaining := maxBuildings - len(selected)
	if remaining > 0 && len(topDirs) > 0 {
		// Add subdirectories from first few top-level directories
		for _, topDir := range topDirs[:min(3, len(topDirs))] {
			if remaining <= 0 {
				break
			}
			for _, subItem := range topDir.Children {
				if remaining <= 0 {
					break
				}
				if subItem.IsDir {
					selected = append(selected, subItem)
					remaining--
				}
			}
		}
	}
	
	// Fill remaining slots with top-level files if needed
	remaining = maxBuildings - len(selected)
	for _, file := range topFiles {
		if remaining <= 0 {
			break
		}
		selected = append(selected, file)
		remaining--
	}
	
	// Now layout the selected buildings with larger footprints
	return layoutBuildings(selected, cols, rows)
}

// layoutBuildings arranges buildings in an Angband-style village with larger multi-glyph buildings
func layoutBuildings(buildings []*domain.FileNode, cols, rows int) []Slot {
	slots := []Slot{}
	
	// Use a grid-based placement with larger building sizes
	buildingSpacing := 2 // Space between buildings
	
	x, y := 1, 1 // Start with margin
	rowHeight := 0
	
	for _, building := range buildings {
		// Determine building size based on type and importance (2x larger)
		var w, h int
		if building.IsDir {
			// Directories are larger buildings (districts)
			w, h = 8, 6
		} else {
			// Files are smaller buildings
			w, h = 6, 4
		}
		
		// Check if building fits on current row
		if x+w > cols-1 {
			// Move to next row
			x = 1
			y += rowHeight + buildingSpacing
			rowHeight = 0
			
			// Check if we're out of vertical space
			if y+h > rows-1 {
				break
			}
		}
		
		// Place the building
		slots = append(slots, Slot{X: x, Y: y, W: w, H: h, Path: building.Path})
		
		// Update position for next building
		x += w + buildingSpacing
		rowHeight = max(rowHeight, h)
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

// BSP Tree structures for natural village layout
type BSPNode struct {
	X, Y, W, H int      // Rectangle bounds
	Left, Right *BSPNode // Child nodes
	IsLeaf     bool      // True if this is a leaf node
	Building   *domain.FileNode // Building placed in this area (for leaves)
	Level      int       // Tree depth level
}

type BSPTree struct {
	Root *BSPNode
	rng  *rand.Rand
}

// NewBSPTree creates a new BSP tree for the given area with deterministic seed
func NewBSPTree(width, height int, seed int64) *BSPTree {
	return &BSPTree{
		Root: &BSPNode{X: 0, Y: 0, W: width, H: height, IsLeaf: true, Level: 0},
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// Split recursively divides the space using BSP algorithm
func (tree *BSPTree) Split(node *BSPNode, maxDepth int, minSize int) {
	if node.Level >= maxDepth || node.W < minSize*2 || node.H < minSize*2 {
		return
	}
	
	node.IsLeaf = false
	
	// Decide split orientation based on aspect ratio
	var splitVertical bool
	ratio := float64(node.W) / float64(node.H)
	
	if ratio > 1.25 {
		splitVertical = true // Wide rectangle, split vertically
	} else if ratio < 0.8 {
		splitVertical = false // Tall rectangle, split horizontally
	} else {
		splitVertical = tree.rng.Float64() > 0.5 // Square-ish, random split
	}
	
	if splitVertical {
		// Vertical split
		minSplit := minSize
		maxSplit := node.W - minSize
		if maxSplit <= minSplit {
			return // Can't split further
		}
		splitPos := minSplit + tree.rng.Intn(maxSplit-minSplit)
		
		node.Left = &BSPNode{
			X: node.X, Y: node.Y, 
			W: splitPos, H: node.H,
			IsLeaf: true, Level: node.Level + 1,
		}
		node.Right = &BSPNode{
			X: node.X + splitPos, Y: node.Y,
			W: node.W - splitPos, H: node.H,
			IsLeaf: true, Level: node.Level + 1,
		}
	} else {
		// Horizontal split
		minSplit := minSize
		maxSplit := node.H - minSize
		if maxSplit <= minSplit {
			return // Can't split further
		}
		splitPos := minSplit + tree.rng.Intn(maxSplit-minSplit)
		
		node.Left = &BSPNode{
			X: node.X, Y: node.Y,
			W: node.W, H: splitPos,
			IsLeaf: true, Level: node.Level + 1,
		}
		node.Right = &BSPNode{
			X: node.X, Y: node.Y + splitPos,
			W: node.W, H: node.H - splitPos,
			IsLeaf: true, Level: node.Level + 1,
		}
	}
	
	// Recursively split children
	tree.Split(node.Left, maxDepth, minSize)
	tree.Split(node.Right, maxDepth, minSize)
}

// GetLeaves returns all leaf nodes from the BSP tree
func (tree *BSPTree) GetLeaves() []*BSPNode {
	var leaves []*BSPNode
	tree.collectLeaves(tree.Root, &leaves)
	return leaves
}

func (tree *BSPTree) collectLeaves(node *BSPNode, leaves *[]*BSPNode) {
	if node == nil {
		return
	}
	if node.IsLeaf {
		*leaves = append(*leaves, node)
	} else {
		tree.collectLeaves(node.Left, leaves)
		tree.collectLeaves(node.Right, leaves)
	}
}

// BSPGrid creates a BSP-based village layout
func BSPGrid(root *domain.FileNode, cols, rows int) []Slot {
	if root == nil {
		return []Slot{}
	}
	
	// Get top-level items (same selection logic as before)
	topLevel := root.Children
	if len(topLevel) == 0 {
		return []Slot{}
	}
	
	var topDirs []*domain.FileNode
	var topFiles []*domain.FileNode
	
	for _, item := range topLevel {
		if item.IsDir {
			topDirs = append(topDirs, item)
		} else {
			topFiles = append(topFiles, item)
		}
	}
	
	sort.Slice(topDirs, func(i, j int) bool { return strings.ToLower(topDirs[i].Name) < strings.ToLower(topDirs[j].Name) })
	sort.Slice(topFiles, func(i, j int) bool { return strings.ToLower(topFiles[i].Name) < strings.ToLower(topFiles[j].Name) })
	
	// Select up to 10 buildings, prioritizing directories
	var selected []*domain.FileNode
	maxBuildings := 10
	
	for _, dir := range topDirs {
		if len(selected) >= maxBuildings {
			break
		}
		selected = append(selected, dir)
	}
	
	remaining := maxBuildings - len(selected)
	if remaining > 0 && len(topDirs) > 0 {
		for _, topDir := range topDirs[:min(3, len(topDirs))] {
			if remaining <= 0 {
				break
			}
			for _, subItem := range topDir.Children {
				if remaining <= 0 {
					break
				}
				if subItem.IsDir {
					selected = append(selected, subItem)
					remaining--
				}
			}
		}
	}
	
	remaining = maxBuildings - len(selected)
	for _, file := range topFiles {
		if remaining <= 0 {
			break
		}
		selected = append(selected, file)
		remaining--
	}
	
	// Create deterministic seed based on directory structure
	seed := int64(0)
	for _, item := range selected {
		seed += int64(Hash(item.Path))
	}
	seed += int64(cols*rows) // Include dimensions in seed
	
	// Create BSP tree and partition the space
	tree := NewBSPTree(cols, rows, seed)
	minRoomSize := 12 // Minimum size for building placement
	maxDepth := 4     // Tree depth
	
	tree.Split(tree.Root, maxDepth, minRoomSize)
	leaves := tree.GetLeaves()
	
	// Place buildings in BSP leaves
	var slots []Slot
	for i, building := range selected {
		if i >= len(leaves) {
			break // More buildings than leaves
		}
		
		leaf := leaves[i]
		
		// Create room within leaf with padding
		padding := 2
		roomW := max(6, leaf.W - padding*2) // Minimum building size
		roomH := max(4, leaf.H - padding*2)
		
		// Adjust building size based on type
		if building.IsDir {
			roomW = min(roomW, 8)
			roomH = min(roomH, 6)
		} else {
			roomW = min(roomW, 6)
			roomH = min(roomH, 4)
		}
		
		// Ensure we have valid ranges for random positioning
		maxOffsetX := leaf.W - roomW - padding*2
		maxOffsetY := leaf.H - roomH - padding*2
		
		roomX := leaf.X + padding
		roomY := leaf.Y + padding
		
		if maxOffsetX > 0 {
			roomX += tree.rng.Intn(maxOffsetX)
		}
		if maxOffsetY > 0 {
			roomY += tree.rng.Intn(maxOffsetY)
		}
		
		slots = append(slots, Slot{
			X: roomX, Y: roomY,
			W: roomW, H: roomH,
			Path: building.Path,
		})
		
		leaf.Building = building
	}
	
	return slots
}

// BSPWithRoads creates a BSP layout with road system
func BSPWithRoads(root *domain.FileNode, cols, rows int) ([]Slot, []Slot) {
	// Get building slots from BSP
	buildingSlots := BSPGrid(root, cols, rows)
	
	// Create deterministic seed for roads based on the same criteria
	seed := int64(0)
	if root != nil {
		for _, child := range root.Children {
			seed += int64(Hash(child.Path))
		}
	}
	seed += int64(cols*rows)
	
	// Create road network connecting the buildings
	tree := NewBSPTree(cols, rows, seed)
	minRoomSize := 12
	maxDepth := 4
	tree.Split(tree.Root, maxDepth, minRoomSize)
	
	roadSlots := generateRoads(tree, cols, rows, buildingSlots)
	
	return buildingSlots, roadSlots
}

// generateRoads creates road network connecting all buildings
func generateRoads(tree *BSPTree, cols, rows int, buildingSlots []Slot) []Slot {
	if len(buildingSlots) == 0 {
		return []Slot{}
	}
	
	// Create road network connecting building centers
	roadCells := make(map[Point]bool)
	
	// Connect all buildings to nearest neighbors using minimum spanning tree approach
	connected := make(map[int]bool)
	connected[0] = true // Start with first building
	
	for len(connected) < len(buildingSlots) {
		minDist := float64(cols * rows) // Large number
		var closestPair [2]int
		
		// Find closest unconnected building to any connected building
		for connectedIdx := range connected {
			for i, building := range buildingSlots {
				if connected[i] {
					continue
				}
				
				connectedBuilding := buildingSlots[connectedIdx]
				dist := distance(
					Point{connectedBuilding.X + connectedBuilding.W/2, connectedBuilding.Y + connectedBuilding.H/2},
					Point{building.X + building.W/2, building.Y + building.H/2},
				)
				
				if dist < minDist {
					minDist = dist
					closestPair = [2]int{connectedIdx, i}
				}
			}
		}
		
		// Connect the closest pair with a road
		if closestPair[1] != 0 || len(connected) == 1 {
			from := buildingSlots[closestPair[0]]
			to := buildingSlots[closestPair[1]]
			
			// Create L-shaped path between building centers
			fromCenter := Point{from.X + from.W/2, from.Y + from.H/2}
			toCenter := Point{to.X + to.W/2, to.Y + to.H/2}
			
			// Horizontal then vertical path
			createPath(roadCells, fromCenter, Point{toCenter.X, fromCenter.Y}, cols, rows)
			createPath(roadCells, Point{toCenter.X, fromCenter.Y}, toCenter, cols, rows)
			
			connected[closestPair[1]] = true
		}
	}
	
	// Convert road cells to road slots
	return cellsToRoadSlots(roadCells)
}

type Point struct {
	X, Y int
}

func distance(a, b Point) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return dx*dx + dy*dy // Using squared distance for efficiency
}

func createPath(roadCells map[Point]bool, from, to Point, cols, rows int) {
	// Create straight line path between two points
	if from.X == to.X {
		// Vertical path
		startY, endY := from.Y, to.Y
		if startY > endY {
			startY, endY = endY, startY
		}
		for y := startY; y <= endY; y++ {
			if from.X >= 0 && from.X < cols && y >= 0 && y < rows {
				roadCells[Point{from.X, y}] = true
			}
		}
	} else {
		// Horizontal path
		startX, endX := from.X, to.X
		if startX > endX {
			startX, endX = endX, startX
		}
		for x := startX; x <= endX; x++ {
			if x >= 0 && x < cols && from.Y >= 0 && from.Y < rows {
				roadCells[Point{x, from.Y}] = true
			}
		}
	}
}

func cellsToRoadSlots(roadCells map[Point]bool) []Slot {
	var roads []Slot
	for point := range roadCells {
		roads = append(roads, Slot{
			X: point.X, Y: point.Y,
			W: 1, H: 1,
			Path: "__road__",
		})
	}
	return roads
}
