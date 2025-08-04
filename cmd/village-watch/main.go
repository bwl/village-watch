// cmd/village-watch/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"example.com/village-watch/internal/config"
	"example.com/village-watch/internal/ui"
)

func main() {
	var path string
	var fps int
	var theme string
	var noUnicode bool
	var ignoreExtra string

	flag.StringVar(&path, "path", ".", "directory to visualize")
	flag.IntVar(&fps, "fps", 20, "target frames per second")
	flag.StringVar(&theme, "theme", "forest", "theme: forest|seaside|desert|contrast")
	flag.BoolVar(&noUnicode, "no-unicode", false, "use ASCII-only tiles")
	flag.StringVar(&ignoreExtra, "ignore", "", "comma-separated ignore globs")
	flag.Parse()

	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	cfg, _ := config.Load(abs) // best-effort load village.yml
	cfg.FPS = fps
	cfg.Theme = theme
	cfg.Render.Unicode = !noUnicode
	cfg.ApplyIgnoreCSV(ignoreExtra)

	m, err := ui.NewModel(abs, cfg)
	if err != nil {
		fmt.Println("init error:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("run error:", err)
		os.Exit(1)
	}
}
