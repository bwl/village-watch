// internal/buildings/buildings.go
package buildings

import (
	"math/rand"
	"strings"

	"example.com/village-watch/internal/domain"
)

// Archetype represents a building category
type Archetype int

const (
	Cottage   Archetype = iota // Code files
	Library                    // Documentation
	Kiosk                      // Config files
	Atelier                    // Assets/media
	Warehouse                  // Binaries/archives
	Academy                    // Test files
	Lantern                    // Log files
	Shrine                     // Special files (.env, secrets)
	District                   // Directories
)

// BuildingDesign defines the visual structure of a building
type BuildingDesign struct {
	Name        string
	Archetype   Archetype
	Unicode     UnicodeDesign
	ASCII       ASCIIDesign
	MinSize     int64 // Minimum file size for this design
	MaxSize     int64 // Maximum file size (0 = no limit)
	Probability float64 // Weight for random selection (0.0-1.0)
}

// UnicodeDesign contains Unicode characters for building parts
type UnicodeDesign struct {
	Corner   rune // Building corners
	Wall     rune // Wall character
	Door     rune // Door/entrance
	Interior rune // Interior fill
	Roof     rune // Optional roof character
}

// ASCIIDesign contains ASCII fallback characters
type ASCIIDesign struct {
	Corner   rune
	Wall     rune
	Door     rune
	Interior rune
	Roof     rune
}

// Registry holds all building designs organized by archetype
type Registry struct {
	designs map[Archetype][]BuildingDesign
}

// NewRegistry creates a new building design registry with default designs
func NewRegistry() *Registry {
	r := &Registry{
		designs: make(map[Archetype][]BuildingDesign),
	}
	r.loadDefaultDesigns()
	return r
}

// GetDesign returns an appropriate design for the given archetype and file size
func (r *Registry) GetDesign(archetype Archetype, size int64, seed int64) BuildingDesign {
	designs := r.designs[archetype]
	if len(designs) == 0 {
		// Fallback to cottage if no designs found
		if archetype != Cottage {
			return r.GetDesign(Cottage, size, seed)
		}
		// Ultimate fallback
		return BuildingDesign{
			Name:      "Basic Building",
			Archetype: Cottage,
			Unicode:   UnicodeDesign{Corner: '▣', Wall: '▬', Door: '▫', Interior: '·', Roof: '⌂'},
			ASCII:     ASCIIDesign{Corner: '+', Wall: '#', Door: '=', Interior: '.', Roof: '^'},
		}
	}

	// Filter designs by size requirements
	var validDesigns []BuildingDesign
	for _, design := range designs {
		if size >= design.MinSize && (design.MaxSize == 0 || size <= design.MaxSize) {
			validDesigns = append(validDesigns, design)
		}
	}

	if len(validDesigns) == 0 {
		validDesigns = designs // Use all if none match size
	}

	// Use deterministic selection based on seed
	rng := rand.New(rand.NewSource(seed))
	
	// Weight-based selection
	totalWeight := 0.0
	for _, design := range validDesigns {
		totalWeight += design.Probability
	}
	
	if totalWeight == 0 {
		return validDesigns[0] // Fallback to first
	}
	
	target := rng.Float64() * totalWeight
	current := 0.0
	
	for _, design := range validDesigns {
		current += design.Probability
		if current >= target {
			return design
		}
	}
	
	return validDesigns[len(validDesigns)-1] // Fallback to last
}

// AddDesign adds a new building design to the registry
func (r *Registry) AddDesign(design BuildingDesign) {
	r.designs[design.Archetype] = append(r.designs[design.Archetype], design)
}

// GetArchetype determines the archetype from a file node
func GetArchetype(n *domain.FileNode) Archetype {
	if n.IsDir {
		return District
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

// loadDefaultDesigns populates the registry with initial building designs
func (r *Registry) loadDefaultDesigns() {
	// Cottage designs (code files)
	r.AddDesign(BuildingDesign{
		Name:        "Simple Cottage",
		Archetype:   Cottage,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '·', Roof: '⌂'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '.', Roof: '^'},
		MinSize:     0,
		MaxSize:     10240, // 10KB
		Probability: 0.6,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Code House",
		Archetype:   Cottage,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '◦', Roof: '⌂'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'o', Roof: '^'},
		MinSize:     10241,
		MaxSize:     131072, // 128KB
		Probability: 0.7,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Code Manor",
		Archetype:   Cottage,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '▣', Roof: '⛪'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '#', Roof: 'M'},
		MinSize:     131073, // 128KB+
		MaxSize:     0,      // No upper limit
		Probability: 0.8,
	})

	// Library designs (documentation)
	r.AddDesign(BuildingDesign{
		Name:        "Small Library",
		Archetype:   Library,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '■', Roof: '📚'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'B', Roof: 'L'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.5,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Grand Library",
		Archetype:   Library,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '▦', Roof: '🏛'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'H', Roof: 'A'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.5,
	})

	// Kiosk designs (config files)
	r.AddDesign(BuildingDesign{
		Name:        "Notice Board",
		Archetype:   Kiosk,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '○', Roof: '📋'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'i', Roof: 'K'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.7,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Config Hall",
		Archetype:   Kiosk,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '⚙', Roof: '⚙'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '*', Roof: 'C'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.3,
	})

	// Atelier designs (art/media files)
	r.AddDesign(BuildingDesign{
		Name:        "Art Studio",
		Archetype:   Atelier,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '◆', Roof: '🎨'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'A', Roof: 'S'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.6,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Gallery",
		Archetype:   Atelier,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '◇', Roof: '🖼'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'o', Roof: 'G'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.4,
	})

	// Warehouse designs (binaries/archives)
	r.AddDesign(BuildingDesign{
		Name:        "Storage Shed",
		Archetype:   Warehouse,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '▤', Roof: '📦'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'S', Roof: 'W'},
		MinSize:     0,
		MaxSize:     1048576, // 1MB
		Probability: 0.6,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Warehouse",
		Archetype:   Warehouse,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '▦', Roof: '🏭'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '#', Roof: 'F'},
		MinSize:     1048577, // 1MB+
		MaxSize:     0,
		Probability: 0.4,
	})

	// Academy designs (test files)
	r.AddDesign(BuildingDesign{
		Name:        "School House",
		Archetype:   Academy,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '⌂', Roof: '🏫'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'T', Roof: 'S'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 1.0,
	})

	// Lantern designs (log files)
	r.AddDesign(BuildingDesign{
		Name:        "Lighthouse",
		Archetype:   Lantern,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '✦', Roof: '💡'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '*', Roof: '!'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 1.0,
	})

	// Shrine designs (secret files)
	r.AddDesign(BuildingDesign{
		Name:        "Sacred Shrine",
		Archetype:   Shrine,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '♦', Roof: '⛩'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '^', Roof: 'T'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 1.0,
	})

	// District designs (directories)
	r.AddDesign(BuildingDesign{
		Name:        "Village Hall",
		Archetype:   District,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '·', Roof: '⌂'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: '.', Roof: 'D'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.4,
	})
	
	r.AddDesign(BuildingDesign{
		Name:        "Grand Hall",
		Archetype:   District,
		Unicode:     UnicodeDesign{Corner: '#', Wall: '#', Door: '=', Interior: '◦', Roof: '🏛'},
		ASCII:       ASCIIDesign{Corner: '#', Wall: '#', Door: '=', Interior: 'o', Roof: 'H'},
		MinSize:     0,
		MaxSize:     0,
		Probability: 0.6,
	})
}