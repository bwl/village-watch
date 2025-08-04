// cmd/village-watch/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"example.com/village-watch/internal/config"
	"example.com/village-watch/internal/scan"
	"example.com/village-watch/internal/scene"
	"example.com/village-watch/internal/ui"
)

func main() {
	var path string
	var fps int
	var theme string
	var noUnicode bool
	var ignoreExtra string
	var testLayout bool

	flag.StringVar(&path, "path", ".", "directory to visualize")
	flag.IntVar(&fps, "fps", 20, "target frames per second")
	flag.StringVar(&theme, "theme", "forest", "theme: forest|seaside|desert|contrast")
	flag.BoolVar(&noUnicode, "no-unicode", false, "use ASCII-only tiles")
	flag.StringVar(&ignoreExtra, "ignore", "", "comma-separated ignore globs")
	flag.BoolVar(&testLayout, "test", false, "test layout generation and print to console")
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

	// Test layout mode - print village layout to console
	if testLayout {
		err := testVillageLayout(abs, &cfg)
		if err != nil {
			fmt.Println("test error:", err)
			os.Exit(1)
		}
		return
	}

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

// testVillageLayout generates and prints the village layout to console
func testVillageLayout(path string, cfg *config.Config) error {
	fmt.Printf("=== Village Layout Test ===\n")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Virtual Map Size: %dx%d\n", scene.VirtualMapWidth, scene.VirtualMapHeight)
	fmt.Printf("Unicode: %v\n", cfg.Render.Unicode)
	fmt.Printf("Theme: %s\n", cfg.Theme)
	fmt.Printf("\n")

	// Scan directory using the correct function
	repo, err := scan.BuildTree(path, *cfg)
	if err != nil {
		return fmt.Errorf("scanning directory: %w", err)
	}

	// Generate scene using virtual map dimensions
	sc := scene.Derive(repo, scene.VirtualMapWidth, scene.VirtualMapHeight, cfg.Render.Unicode)

	fmt.Printf("Generated village with %d buildings:\n", len(repo.Index)-1)
	
	// Print file list
	fmt.Printf("\nFiles/Directories found:\n")
	for path, node := range repo.Index {
		if path == repo.Root.Path {
			continue // Skip root
		}
		nodeType := "FILE"
		if node.IsDir {
			nodeType = "DIR "
		}
		fmt.Printf("  %s %s\n", nodeType, strings.TrimPrefix(path, repo.Root.Path+"/"))
	}

	fmt.Printf("\n=== Village Map ===\n")
	
	// Print the entire virtual map
	for _, line := range sc.Canvas {
		fmt.Println(line)
	}
	
	fmt.Printf("\n%s\n", sc.Status)
	fmt.Printf("=== End Map ===\n")

	return nil
}
