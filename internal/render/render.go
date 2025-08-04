// internal/render/render.go
package render

import (
	"example.com/village-watch/internal/scene"
	lg "github.com/charmbracelet/lipgloss"
	"strings"
)

type Theme struct {
	Ground lg.Style
	HUD    lg.Style
}

func ThemeByName(name string) Theme {
	switch name {
	case "seaside":
		return Theme{Ground: lg.NewStyle().Foreground(lg.Color("87")), HUD: lg.NewStyle().Foreground(lg.Color("81"))}
	case "desert":
		return Theme{Ground: lg.NewStyle().Foreground(lg.Color("179")), HUD: lg.NewStyle().Foreground(lg.Color("180"))}
	case "contrast":
		return Theme{Ground: lg.NewStyle().Foreground(lg.Color("15")).Background(lg.Color("0")), HUD: lg.NewStyle().Foreground(lg.Color("15")).Background(lg.Color("0"))}
	default:
		return Theme{Ground: lg.NewStyle().Foreground(lg.Color("120")), HUD: lg.NewStyle().Foreground(lg.Color("108"))}
	}
}

func View(sc scene.Scene, t Theme, width, height int) string {
	return ViewWithStatus(sc, t, width, height, false, false, "forest")
}

func ViewWithStatus(sc scene.Scene, t Theme, width, height int, paused, filterActive bool, currentTheme string) string {
	// Draw canvas into available height-2 (reserve status bar and help line)
	maxRows := height - 2
	if maxRows < 1 {
		maxRows = 1
	}
	lines := sc.Canvas
	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}
	b := strings.Builder{}
	for _, ln := range lines {
		b.WriteString(t.Ground.Render(ln))
		b.WriteByte('\n')
	}
	
	// Enhanced status line with indicators
	statusLine := sc.Status
	if paused {
		statusLine = "[PAUSED] " + statusLine
	}
	if filterActive {
		statusLine = "[FILTERED] " + statusLine
	}
	statusLine += " | Theme: " + currentTheme
	
	b.WriteString(t.HUD.Render(statusLine))
	b.WriteByte('\n')
	b.WriteString(t.HUD.Render("(q) quit  (p) pause  (h) help  (f) filter  (t) theme  (r) refresh"))
	return b.String()
}

func ViewWithHelp(sc scene.Scene, t Theme, width, height int, paused, filterActive bool, currentTheme string) string {
	helpText := []string{
		"Village Watch - Filesystem Visualizer",
		"",
		"Controls:",
		"  q / Ctrl+C  - Quit application",
		"  p           - Pause/unpause updates",
		"  h           - Toggle this help",
		"  f           - Toggle activity filter",
		"  t           - Cycle themes (forest/seaside/desert/contrast)",
		"  r           - Force refresh filesystem",
		"  Escape      - Close overlays",
		"",
		"Building Types:",
		"  h/H/M - Cottages (Code files: Go, JS, Python, etc.)",
		"  L     - Libraries (Documentation: MD, RST, TXT)",
		"  K     - Kiosks (Config files: YAML, JSON, TOML)",
		"  A     - Ateliers (Assets: PNG, JPG, SVG)",
		"  W     - Warehouses (Archives: ZIP, TAR, binaries)",
		"  S     - Academy (Test files)",
		"  *     - Lanterns (Log files)",
		"  ^     - Shrines (Secret files: .env, passwords)",
		"",
		"Animations:",
		"  +     - Construction (New files being created)",
		"  ~     - Activity smoke (Files being modified)",
		"  X     - Demolition (Files being deleted)",
		"",
		"Press any key to continue...",
	}
	
	b := strings.Builder{}
	for i, line := range helpText {
		if i < height-1 {
			b.WriteString(t.HUD.Render(line))
			b.WriteByte('\n')
		}
	}
	return b.String()
}
