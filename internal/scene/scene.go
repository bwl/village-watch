// internal/scene/scene.go
package scene

import (
	"fmt"
	"strings"

	"example.com/village-watch/internal/domain"
	"example.com/village-watch/internal/layout"
)

type Tile struct{ Glyph string }

type Scene struct {
	Canvas []string
	W, H   int
	Status string
}

func Derive(repo *domain.RepoState, cols, rows int, unicode bool) Scene {
	return DeriveWithFPS(repo, cols, rows, unicode, 0)
}

func DeriveWithFPS(repo *domain.RepoState, cols, rows int, unicode bool, fps float64) Scene {
	if repo == nil || repo.Root == nil {
		return Scene{Canvas: []string{"(empty)"}}
	}
	grid := make([][]rune, rows)
	for i := range grid {
		grid[i] = make([]rune, cols)
		for j := 0; j < cols; j++ {
			if unicode {
				grid[i][j] = '░' // Light grass/ground
			} else {
				grid[i][j] = '.'
			}
		}
	}
	
	slots := layout.Grid(repo.Root, cols, rows)
	for _, s := range slots {
		renderBuilding(grid, repo, s, cols, rows, unicode)
	}
	
	lines := make([]string, rows)
	for y := 0; y < rows; y++ {
		lines[y] = string(grid[y])
	}
	
	// Add roads/paths between districts
	addRoads(grid, slots, cols, rows, unicode)
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
	
	return Scene{Canvas: lines, W: cols, H: rows, Status: status}
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

// drawDistrict renders a district (directory) with gates and nameplate
func drawDistrict(grid [][]rune, slot layout.Slot, node *domain.FileNode, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	if unicode {
		// District boundary with gate
		for dx := 0; dx < w && x+dx < cols; dx++ {
			for dy := 0; dy < h && y+dy < rows; dy++ {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					if dx == w/2 && dy == 0 {
						grid[y+dy][x+dx] = '⌬' // Gate
					} else {
						grid[y+dy][x+dx] = '▢' // District wall
					}
				} else {
					grid[y+dy][x+dx] = '⌂' // Houses inside district
				}
			}
		}
	} else {
		// ASCII district
		for dx := 0; dx < w && x+dx < cols; dx++ {
			for dy := 0; dy < h && y+dy < rows; dy++ {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					if dx == w/2 && dy == 0 {
						grid[y+dy][x+dx] = '^' // Gate
					} else {
						grid[y+dy][x+dx] = '#' // District wall
					}
				} else {
					grid[y+dy][x+dx] = 'H' // Houses inside district
				}
			}
		}
	}
}

// drawBuilding renders a single building (file)
func drawBuilding(grid [][]rune, slot layout.Slot, node *domain.FileNode, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	if x >= cols || y >= rows {
		return
	}
	
	archetype := getArchetype(node)
	glyph := glyphForArchetype(archetype, node.Size, unicode)
	grid[y][x] = glyph
}

// drawAnimationEffect renders animation states
func drawAnimationEffect(grid [][]rune, slot layout.Slot, state domain.FileState, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	if x >= cols || y >= rows {
		return
	}
	
	glyph := glyphForState(state, unicode)
	grid[y][x] = glyph
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
