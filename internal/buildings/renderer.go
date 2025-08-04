// internal/buildings/renderer.go
package buildings

import (
	"hash/fnv"

	"example.com/village-watch/internal/domain"
	"example.com/village-watch/internal/layout"
)

// Renderer handles drawing buildings using the design registry
type Renderer struct {
	registry *Registry
}

func (r *Renderer) RenderLabel(grid [][]rune, slot layout.Slot, name string, cols, rows int) {
	x, y := slot.X, slot.Y
	w := slot.W
	if w <= 0 || len(name) == 0 { return }
	runes := []rune(name)
	max := w
	if max <= 0 { return }
	if len(runes) > max {
		if max >= 2 {
			runes = append(runes[:max-1], '…')
		} else {
			runes = runes[:max]
		}
	}
	startX := x + (w-len(runes))/2
	labelY := y - 1
	if labelY < 0 { labelY = y }
	for i := 0; i < len(runes); i++ {
		px := startX + i
		if px >= 0 && px < cols && labelY >= 0 && labelY < rows {
			grid[labelY][px] = runes[i]
		}
	}
}

// NewRenderer creates a new building renderer with default designs
func NewRenderer() *Renderer {
	return &Renderer{
		registry: NewRegistry(),
	}
}

// GetRegistry returns the building design registry for customization
func (r *Renderer) GetRegistry() *Registry {
	return r.registry
}

// RenderBuilding draws a building at the given slot using the appropriate design
func (r *Renderer) RenderBuilding(grid [][]rune, repo *domain.RepoState, slot layout.Slot, cols, rows int, unicode bool) {
	node := repo.Index[slot.Path]
	if node == nil {
		return
	}
	
	// Handle animation states first
	if node.IsStateActive() {
		r.drawAnimationEffect(grid, slot, node.State, cols, rows, unicode)
		return
	}
	
	// Get appropriate design
	archetype := GetArchetype(node)
	seed := r.generateSeed(node.Path, node.Size)
	design := r.registry.GetDesign(archetype, node.Size, seed)
	
	// Render the building using the design
	if node.IsDir {
		r.drawDistrictBuilding(grid, slot, design, cols, rows, unicode)
	} else {
		r.drawFileBuilding(grid, slot, design, cols, rows, unicode)
	}
}

// drawFileBuilding renders a file as a building using the specified design
func (r *Renderer) drawFileBuilding(grid [][]rune, slot layout.Slot, design BuildingDesign, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	var corner, wall, door, interior rune
	
	if unicode {
		corner = design.Unicode.Corner
		wall = design.Unicode.Wall
		door = design.Unicode.Door
		interior = design.Unicode.Interior
	} else {
		corner = design.ASCII.Corner
		wall = design.ASCII.Wall
		door = design.ASCII.Door
		interior = design.ASCII.Interior
	}
	
	// Draw building structure
	for dx := 0; dx < w && x+dx < cols; dx++ {
		for dy := 0; dy < h && y+dy < rows; dy++ {
			if x+dx >= 0 && y+dy >= 0 {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					// Building outline
					if (dx == 0 || dx == w-1) && (dy == 0 || dy == h-1) {
						grid[y+dy][x+dx] = corner // Corners
					} else {
						grid[y+dy][x+dx] = wall // Walls
					}
					// Door in the middle of bottom wall
					if dy == h-1 && dx == w/2 {
						grid[y+dy][x+dx] = door
					}
				} else {
					// Interior
					grid[y+dy][x+dx] = interior
				}
			}
		}
	}
}

// drawDistrictBuilding renders a directory as a district building
func (r *Renderer) drawDistrictBuilding(grid [][]rune, slot layout.Slot, design BuildingDesign, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	var corner, wall, door, interior, roof rune
	
	if unicode {
		corner = design.Unicode.Corner
		wall = design.Unicode.Wall
		door = design.Unicode.Door
		interior = design.Unicode.Interior
		roof = design.Unicode.Roof
	} else {
		corner = design.ASCII.Corner
		wall = design.ASCII.Wall
		door = design.ASCII.Door
		interior = design.ASCII.Interior
		roof = design.ASCII.Roof
	}
	
	// Draw district as a complex building with internal structure
	for dx := 0; dx < w && x+dx < cols; dx++ {
		for dy := 0; dy < h && y+dy < rows; dy++ {
			if x+dx >= 0 && y+dy >= 0 {
				if dx == 0 || dx == w-1 || dy == 0 || dy == h-1 {
					// Building outline
					if (dx == 0 || dx == w-1) && (dy == 0 || dy == h-1) {
						grid[y+dy][x+dx] = corner // Corners
					} else {
						grid[y+dy][x+dx] = wall // Walls
					}
					// Door in the middle of bottom wall
					if dy == h-1 && dx == w/2 {
						grid[y+dy][x+dx] = door
					}
				} else {
					// Interior - show different room features for districts
					if dx == 1 && dy == 1 {
						grid[y+dy][x+dx] = roof // Main feature in top-left
					} else if dx == w-2 && dy == 1 && w > 3 {
						grid[y+dy][x+dx] = corner // Secondary feature in top-right
					} else {
						grid[y+dy][x+dx] = interior // Floor
					}
				}
			}
		}
	}
}

// drawAnimationEffect renders animation states across the entire building
func (r *Renderer) drawAnimationEffect(grid [][]rune, slot layout.Slot, state domain.FileState, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	// Get animation glyph
	var animGlyph rune
	switch state {
	case domain.StateNew:
		animGlyph = '+' // Construction
	case domain.StateModified:
		animGlyph = '~' // Activity/smoke
	case domain.StateDeleted:
		animGlyph = 'X' // Demolished
	default:
		animGlyph = '?'
	}
	
	// Fill the entire building area based on animation type
	for dx := 0; dx < w && x+dx < cols; dx++ {
		for dy := 0; dy < h && y+dy < rows; dy++ {
			if x+dx >= 0 && y+dy >= 0 {
				// For construction/demolition, show effect throughout
				if state == domain.StateNew || state == domain.StateDeleted {
					grid[y+dy][x+dx] = animGlyph
				} else if state == domain.StateModified {
					// For modification, show smoke/activity effects around building perimeter
					if dx == 0 || dx == w-1 || dy == 0 {
						grid[y+dy][x+dx] = animGlyph
					}
				}
			}
		}
	}
}

// generateSeed creates a deterministic seed for design selection
func (r *Renderer) generateSeed(path string, size int64) int64 {
	h := fnv.New64a()
	h.Write([]byte(path))
	return int64(h.Sum64()) + size
}

// RenderRoad draws a road segment
func (r *Renderer) RenderRoad(grid [][]rune, slot layout.Slot, cols, rows int, unicode bool) {
	x, y := slot.X, slot.Y
	w, h := slot.W, slot.H
	
	roadGlyph := '·'
	if unicode {
		roadGlyph = '▫' // Light road glyph
	}
	
	// Fill the road area
	for dx := 0; dx < w && x+dx < cols; dx++ {
		for dy := 0; dy < h && y+dy < rows; dy++ {
			if x+dx >= 0 && y+dy >= 0 && x+dx < cols && y+dy < rows {
				grid[y+dy][x+dx] = roadGlyph
			}
		}
	}
}