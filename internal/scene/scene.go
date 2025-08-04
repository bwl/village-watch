// internal/scene/scene.go
package scene

import (
	"fmt"
	"strings"

	"example.com/village-watch/internal/domain"
	"example.com/village-watch/internal/layout"
)

type Tile struct{ Glyph string }

const (
	VirtualMapWidth  = 128
	VirtualMapHeight = 60
)

type Scene struct {
	Canvas []string
	W, H   int
	Status string
	VirtualMap [][]rune // Full 256x224 virtual map
	ViewportX, ViewportY int // Current viewport position
}

func Derive(repo *domain.RepoState, cols, rows int, unicode bool) Scene {
	return DeriveWithFPS(repo, cols, rows, unicode, 0)
}

func DeriveWithFPS(repo *domain.RepoState, cols, rows int, unicode bool, fps float64) Scene {
	if repo == nil || repo.Root == nil {
		return Scene{Canvas: []string{"(empty)"}}
	}
	
	// Create virtual map (always 256x224)
	virtualMap := make([][]rune, VirtualMapHeight)
	for i := range virtualMap {
		virtualMap[i] = make([]rune, VirtualMapWidth)
		for j := 0; j < VirtualMapWidth; j++ {
			if unicode {
				virtualMap[i][j] = '░' // Light grass/ground
			} else {
				virtualMap[i][j] = '.'
			}
		}
	}
	
	// Generate layout on virtual map dimensions
	buildingSlots, roadSlots := layout.BSPWithRoads(repo.Root, VirtualMapWidth, VirtualMapHeight)
	
	// Render roads first (so buildings can overlap them)
	for _, road := range roadSlots {
		renderRoad(virtualMap, road, VirtualMapWidth, VirtualMapHeight, unicode)
	}
	
	// Then render buildings
	for _, s := range buildingSlots {
		renderBuilding(virtualMap, repo, s, VirtualMapWidth, VirtualMapHeight, unicode)
	}
	
	// Create viewport of the virtual map
	viewportX, viewportY := calculateViewport(cols, rows)
	canvas := extractViewport(virtualMap, cols, rows, viewportX, viewportY)
	
	// Count active animations
	animCount := 0
	for _, node := range repo.Index {
		if node.IsStateActive() {
			animCount++
		}
	}
	
	// Rich status bar with village information
	var status string
	if fps > 0 {
		status = fmt.Sprintf("Village: %d buildings | New: %d | Modified: %d | Deleted: %d | Animations: %d | FPS: %.1f", 
			len(repo.Index)-1, // -1 to exclude root dir
			repo.Stats.NewFiles, 
			repo.Stats.Modified, 
			repo.Stats.Deleted,
			animCount,
			fps)
	} else {
		status = fmt.Sprintf("Village: %d buildings | New: %d | Modified: %d | Deleted: %d | Animations: %d", 
			len(repo.Index)-1, // -1 to exclude root dir
			repo.Stats.NewFiles, 
			repo.Stats.Modified, 
			repo.Stats.Deleted,
			animCount)
	}
	
	return Scene{
		Canvas: canvas, 
		W: cols, H: rows, 
		Status: status,
		VirtualMap: virtualMap,
		ViewportX: viewportX, ViewportY: viewportY,
	}
}

func glyphFor(repo *domain.RepoState, path string, unicode bool) rune {
	n := repo.Index[path]
	if n == nil {
		return '?'
	}
	
	// Handle animation states first
	if n.IsStateActive() {
		return glyphForState(n.State, unicode)
	}
	
	if n.IsDir {
		// Districts with gates
		if unicode {
			return '⌂' // House for districts
		}
		return '#'
	}
	
	// Determine building archetype based on file type
	archetype := getArchetype(n)
	return glyphForArchetype(archetype, n.Size, unicode)
}

func glyphForState(state domain.FileState, unicode bool) rune {
	// Always use ASCII animation states
	switch state {
	case domain.StateNew:
		return '+' // Construction
	case domain.StateModified:
		return '~' // Activity/smoke
	case domain.StateDeleted:
		return 'X' // Demolished
	default:
		return '?'
	}
}

type Archetype int

const (
	Cottage Archetype = iota // Code files
	Library                  // Documentation
	Kiosk                    // Config files
	Atelier                  // Assets/media
	Warehouse                // Binaries/archives
	Academy                  // Test files
	Lantern                  // Log files
	Shrine                   // Special files (.env, secrets)
)

func getArchetype(n *domain.FileNode) Archetype {
	if n.IsDir {
		return Cottage // shouldn't happen here but safe default
	}
	
	ext := strings.ToLower(n.Ext)
	name := strings.ToLower(n.Name)
	
	// Log files
	if ext == ".log" || strings.Contains(name, "log") {
		return Lantern
	}
	
	// Test files
	if strings.Contains(name, "test") || strings.Contains(name, "_test") || strings.Contains(name, ".test") {
		return Academy
	}
	
	// Secret/config files
	if name == ".env" || strings.Contains(name, "secret") || strings.Contains(name, "password") {
		return Shrine
	}
	
	// File type mappings
	switch ext {
	case ".md", ".rst", ".txt", ".doc", ".pdf":
		return Library
	case ".yaml", ".yml", ".json", ".toml", ".ini", ".conf", ".config":
		return Kiosk
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp":
		return Atelier
	case ".zip", ".tar", ".gz", ".bz2", ".rar", ".7z", ".exe", ".dll", ".so", ".dylib", ".a":
		return Warehouse
	case ".go", ".js", ".ts", ".tsx", ".py", ".rs", ".java", ".c", ".cpp", ".h", ".cs", ".php", ".rb", ".swift":
		return Cottage
	default:
		return Cottage // Default to cottage for unknown files
	}
}

func glyphForArchetype(arch Archetype, size int64, unicode bool) rune {
	// Size tiers for building scale
	isLarge := size > 128*1024 // 128KB+
	isHuge := size > 2*1024*1024 // 2MB+
	
	if unicode {
		switch arch {
		case Cottage:
			if isHuge {
				return '▲' // Manor
			} else if isLarge {
				return '▲' // House
			}
			return '△' // cottage
		case Library:
			return '■' // Library
		case Kiosk:
			return '▢' // Notice board
		case Atelier:
			return '◆' // Art studio
		case Warehouse:
			if isLarge {
				return '▦' // Factory
			}
			return '▤' // Warehouse
		case Academy:
			return '⌂' // School
		case Lantern:
			return '✦' // Lantern
		case Shrine:
			return '♦' // Shrine
		default:
			return '●'
		}
	} else {
		// ASCII fallbacks
		switch arch {
		case Cottage:
			if isHuge {
				return 'M' // Manor
			} else if isLarge {
				return 'H' // House
			}
			return 'h' // cottage
		case Library:
			return 'L'
		case Kiosk:
			return 'K'
		case Atelier:
			return 'A'
		case Warehouse:
			return 'W'
		case Academy:
			return 'S' // School
		case Lantern:
			return '*' // Lantern/fire
		case Shrine:
			return '^' // Shrine
		default:
			return 'f' // file
		}
	}
}

// renderBuilding draws a building or district at the given slot
func renderBuilding(grid [][]rune, repo *domain.RepoState, slot layout.Slot, cols, rows int, unicode bool) {
	node := repo.Index[slot.Path]
	if node == nil {
		return
	}
	
	// Handle animation states first
	if node.IsStateActive() {
		drawAnimationEffect(grid, slot, node.State, cols, rows, unicode)
		return
	}
	
	if node.IsDir {
		drawDistrict(grid, slot, node, cols, rows, unicode)
	} else {
		drawBuilding(grid, slot, node, cols, rows, unicode)
	}
}

// drawDistrict renders a district (directory) as a larger Angband-style building
func drawDistrict(grid [][]rune, slot layout.Slot, node *domain.FileNode, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	if unicode {
		// Draw district as a multi-room building with internal structure
		for dx := 0; dx < w && x+dx < cols; dx++ {
			for dy := 0; dy < h && y+dy < rows; dy++ {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					// Walls
					if (dx == 0 || dx == w-1) && (dy == 0 || dy == h-1) {
						grid[y+dy][x+dx] = '▣' // Corner
					} else if dx == 0 || dx == w-1 {
						grid[y+dy][x+dx] = '▮' // Vertical wall
					} else {
						grid[y+dy][x+dx] = '▬' // Horizontal wall
					}
					// Door in the middle of bottom wall
					if dy == h-1 && dx == w/2 {
						grid[y+dy][x+dx] = '▫' // Door
					}
				} else {
					// Interior - show different room types
					if dx == 1 && dy == 1 {
						grid[y+dy][x+dx] = '⌂' // Main room
					} else if dx == w-2 && dy == 1 && w > 3 {
						grid[y+dy][x+dx] = '◊' // Side room
					} else {
						grid[y+dy][x+dx] = '·' // Floor
					}
				}
			}
		}
	} else {
		// ASCII district - structured building
		for dx := 0; dx < w && x+dx < cols; dx++ {
			for dy := 0; dy < h && y+dy < rows; dy++ {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					// Walls with corners
					if (dx == 0 || dx == w-1) && (dy == 0 || dy == h-1) {
						grid[y+dy][x+dx] = '+' // Corner
					} else {
						grid[y+dy][x+dx] = '#' // Wall
					}
					// Door
					if dy == h-1 && dx == w/2 {
						grid[y+dy][x+dx] = '=' // Door
					}
				} else {
					// Interior rooms
					if dx == 1 && dy == 1 {
						grid[y+dy][x+dx] = 'D' // Directory marker
					} else if dx == w-2 && dy == 1 && w > 3 {
						grid[y+dy][x+dx] = 'o' // Sub-room
					} else {
						grid[y+dy][x+dx] = '.' // Floor
					}
				}
			}
		}
	}
}

// drawBuilding renders a single building (file) as a multi-glyph structure
func drawBuilding(grid [][]rune, slot layout.Slot, node *domain.FileNode, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	archetype := getArchetype(node)
	
	if unicode {
		// Multi-glyph Unicode buildings
		for dx := 0; dx < w && x+dx < cols; dx++ {
			for dy := 0; dy < h && y+dy < rows; dy++ {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					// Building outline
					if (dx == 0 || dx == w-1) && (dy == 0 || dy == h-1) {
						grid[y+dy][x+dx] = '▪' // Corner
					} else {
						grid[y+dy][x+dx] = getBuildingWall(archetype, unicode)
					}
					// Door
					if dy == h-1 && dx == w/2 {
						grid[y+dy][x+dx] = '▫'
					}
				} else {
					// Interior based on archetype
					grid[y+dy][x+dx] = getBuildingInterior(archetype, node.Size, unicode)
				}
			}
		}
	} else {
		// ASCII multi-glyph buildings
		for dx := 0; dx < w && x+dx < cols; dx++ {
			for dy := 0; dy < h && y+dy < rows; dy++ {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					// Building outline
					if (dx == 0 || dx == w-1) && (dy == 0 || dy == h-1) {
						grid[y+dy][x+dx] = '+'
					} else {
						grid[y+dy][x+dx] = getBuildingWall(archetype, unicode)
					}
					// Door
					if dy == h-1 && dx == w/2 {
						grid[y+dy][x+dx] = '='
					}
				} else {
					// Interior
					grid[y+dy][x+dx] = getBuildingInterior(archetype, node.Size, unicode)
				}
			}
		}
	}
}

// getBuildingWall returns appropriate wall character for building type
func getBuildingWall(arch Archetype, unicode bool) rune {
	if unicode {
		switch arch {
		case Library:
			return '▬' // Solid wall for libraries
		case Kiosk:
			return '▢' // Light wall for kiosks
		case Atelier:
			return '▥' // Dotted wall for ateliers
		case Warehouse:
			return '▦' // Heavy wall for warehouses
		case Academy:
			return '▤' // Patterned wall for schools
		case Lantern:
			return '▨' // Diagonal wall for lanterns
		case Shrine:
			return '▩' // Dark wall for shrines
		default:
			return '▬' // Default wall
		}
	} else {
		return '#' // ASCII wall
	}
}

// getBuildingInterior returns interior character based on archetype and size
func getBuildingInterior(arch Archetype, size int64, unicode bool) rune {
	if unicode {
		switch arch {
		case Library:
			return '■' // Books/shelves
		case Kiosk:
			return '○' // Notice board
		case Atelier:
			return '◆' // Art supplies
		case Warehouse:
			if size > 1024*1024 {
				return '▦' // Heavy storage
			}
			return '▤' // Light storage
		case Academy:
			return '⌂' // Desks
		case Lantern:
			return '✦' // Light source
		case Shrine:
			return '♦' // Sacred item
		default:
			return '·' // Floor
		}
	} else {
		// ASCII interiors
		switch arch {
		case Library:
			return 'B' // Books
		case Kiosk:
			return 'i' // Info
		case Atelier:
			return 'A' // Art
		case Warehouse:
			return 'S' // Storage
		case Academy:
			return 'T' // Teaching
		case Lantern:
			return '*' // Light
		case Shrine:
			return '^' // Sacred
		default:
			return '.' // Floor
		}
	}
}

// drawAnimationEffect renders animation states across the entire building
func drawAnimationEffect(grid [][]rune, slot layout.Slot, state domain.FileState, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	// Fill the entire building area with animation effect
	animGlyph := glyphForState(state, unicode)
	
	for dx := 0; dx < w && x+dx < cols; dx++ {
		for dy := 0; dy < h && y+dy < rows; dy++ {
			// For construction/demolition, show effect throughout
			if state == domain.StateNew || state == domain.StateDeleted {
				grid[y+dy][x+dx] = animGlyph
			} else if state == domain.StateModified {
				// For modification, show smoke/activity effects around building
				if dx == 0 || dx == w-1 || dy == 0 {
					grid[y+dy][x+dx] = animGlyph
				}
			}
		}
	}
}

// renderRoad draws a road segment
func renderRoad(grid [][]rune, slot layout.Slot, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	roadGlyph := '·'
	if unicode {
		roadGlyph = '▫' // Light road glyph
	}
	
	// Fill the road area
	for dx := 0; dx < w && x+dx < cols; dx++ {
		for dy := 0; dy < h && y+dy < rows; dy++ {
			if x+dx >= 0 && y+dy >= 0 {
				grid[y+dy][x+dx] = roadGlyph
			}
		}
	}
}

// addRoads draws paths between districts and major buildings
func addRoads(grid [][]rune, slots []layout.Slot, cols, rows int, unicode bool) {
	roadGlyph := '·'
	if unicode {
		roadGlyph = '▫'
	}
	
	// Simple horizontal roads every few rows
	for y := 3; y < rows; y += 4 {
		for x := 0; x < cols; x++ {
			if grid[y][x] == '░' || grid[y][x] == '.' {
				grid[y][x] = roadGlyph
			}
		}
	}
}

// calculateViewport determines where to position the viewport on the virtual map
func calculateViewport(viewWidth, viewHeight int) (int, int) {
	// If viewport is larger than virtual map, center the virtual map
	if viewWidth >= VirtualMapWidth && viewHeight >= VirtualMapHeight {
		return 0, 0 // Show entire virtual map
	}
	
	// If viewport is smaller, center it on the virtual map
	viewportX := (VirtualMapWidth - viewWidth) / 2
	viewportY := (VirtualMapHeight - viewHeight) / 2
	
	// Ensure viewport doesn't go negative
	if viewportX < 0 {
		viewportX = 0
	}
	if viewportY < 0 {
		viewportY = 0
	}
	
	return viewportX, viewportY
}

// extractViewport extracts a viewport from the virtual map
func extractViewport(virtualMap [][]rune, viewWidth, viewHeight, viewportX, viewportY int) []string {
	canvas := make([]string, viewHeight)
	
	for y := 0; y < viewHeight; y++ {
		line := make([]rune, viewWidth)
		for x := 0; x < viewWidth; x++ {
			mapX := viewportX + x
			mapY := viewportY + y
			
			// If viewport is larger than virtual map, show virtual map centered with padding
			if viewWidth > VirtualMapWidth || viewHeight > VirtualMapHeight {
				// Calculate centering offsets
				offsetX := (viewWidth - VirtualMapWidth) / 2
				offsetY := (viewHeight - VirtualMapHeight) / 2
				
				if x >= offsetX && x < offsetX+VirtualMapWidth && 
				   y >= offsetY && y < offsetY+VirtualMapHeight {
					// Inside virtual map bounds
					virtualX := x - offsetX
					virtualY := y - offsetY
					if virtualX >= 0 && virtualX < VirtualMapWidth && 
					   virtualY >= 0 && virtualY < VirtualMapHeight {
						line[x] = virtualMap[virtualY][virtualX]
					} else {
						line[x] = ' ' // Padding
					}
				} else {
					line[x] = ' ' // Padding around virtual map
				}
			} else {
				// Normal viewport (smaller than virtual map)
				if mapX >= 0 && mapX < VirtualMapWidth && 
				   mapY >= 0 && mapY < VirtualMapHeight {
					line[x] = virtualMap[mapY][mapX]
				} else {
					line[x] = ' ' // Outside virtual map bounds
				}
			}
		}
		canvas[y] = string(line)
	}
	
	return canvas
}
