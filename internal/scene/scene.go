// internal/scene/scene.go
package scene

import (
	"fmt"

	"example.com/village-watch/internal/buildings"
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
	VirtualMap [][]rune // Full 128x60 virtual map
	ViewportX, ViewportY int // Current viewport position
	buildingRenderer *buildings.Renderer // Modular building renderer
}

func Derive(repo *domain.RepoState, cols, rows int, unicode bool) Scene {
	return DeriveWithFPS(repo, cols, rows, unicode, 0)
}

// GetBuildingRenderer returns the building renderer for customization
func (s *Scene) GetBuildingRenderer() *buildings.Renderer {
	return s.buildingRenderer
}

func DeriveWithFPS(repo *domain.RepoState, cols, rows int, unicode bool, fps float64) Scene {
	if repo == nil || repo.Root == nil {
		return Scene{Canvas: []string{"(empty)"}}
	}
	
	// Create building renderer
	buildingRenderer := buildings.NewRenderer()
	
	// Create virtual map (always 128x60)
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
		buildingRenderer.RenderRoad(virtualMap, road, VirtualMapWidth, VirtualMapHeight, unicode)
	}
	
	// Then render buildings using new modular system
	for _, s := range buildingSlots {
		buildingRenderer.RenderBuilding(virtualMap, repo, s, VirtualMapWidth, VirtualMapHeight, unicode)
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
		buildingRenderer: buildingRenderer,
	}
}

// addRoads draws paths between districts and major buildings (legacy function - kept for compatibility)
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
