// internal/ui/model.go
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"example.com/village-watch/internal/config"
	"example.com/village-watch/internal/domain"
	"example.com/village-watch/internal/render"
	"example.com/village-watch/internal/scan"
	"example.com/village-watch/internal/scene"
	"example.com/village-watch/internal/watch"
)

type tickMsg time.Time
type eventsMsg watch.EventOut

type Model struct {
	root          string
	cfg           config.Config
	width, height int
	repo          *domain.RepoState
	scene         scene.Scene
	paused        bool
	out           chan watch.EventOut
	stop          func() error
	lastResize    time.Time
	lastTick      time.Time
	fps           float64
	frameCount    int
	showHelp      bool
	filterActive  bool
}

func NewModel(root string, cfg config.Config) (Model, error) {
	repo, err := scan.BuildTree(root, cfg)
	if err != nil {
		return Model{}, err
	}
	out, stop, err := watch.Start(root, cfg)
	if err != nil {
		return Model{}, err
	}
	m := Model{root: root, cfg: cfg, repo: repo, out: out, stop: stop}
	return m, nil
}

func (m Model) Init() tea.Cmd { return tea.Batch(tick(m.cfg.FPS), waitEvents(m.out)) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.stop != nil {
				_ = m.stop()
			}
			return m, tea.Quit
		case "p":
			m.paused = !m.paused
		case "h":
			m.showHelp = !m.showHelp
		case "f":
			m.filterActive = !m.filterActive
		case "t":
			// Cycle through themes
			themes := []string{"forest", "seaside", "desert", "contrast"}
			for i, theme := range themes {
				if m.cfg.Theme == theme {
					m.cfg.Theme = themes[(i+1)%len(themes)]
					break
				}
			}
		case "r":
			// Force refresh
			repo, _ := scan.BuildTree(m.root, m.cfg)
			m.preserveAnimationStates(repo)
			m.repo = repo
		case "escape":
			m.showHelp = false
			m.filterActive = false
		}
		return m, nil
	case tickMsg:
		// Calculate FPS
		now := time.Time(msg)
		if !m.lastTick.IsZero() {
			dt := now.Sub(m.lastTick).Seconds()
			if dt > 0 {
				m.fps = 0.9*m.fps + 0.1*(1.0/dt) // Smooth FPS calculation
			}
		}
		m.lastTick = now
		m.frameCount++
		
		if !m.paused {
			// Update animation states (expire old ones)
			m.repo.UpdateStates()
			m.scene = scene.DeriveWithFPS(m.repo, max(10, m.width), max(5, m.height-2), m.cfg.Render.Unicode, m.fps)
		}
		return m, tick(m.cfg.FPS)
	case eventsMsg:
		// Process events and set animation states
		for _, e := range msg.Events {
			switch e.Kind {
			case domain.Create:
				m.repo.Stats.NewFiles++
				// Mark new files with construction animation for 2 seconds
				m.repo.SetFileState(e.Path, domain.StateNew, 2*time.Second)
			case domain.Write:
				m.repo.Stats.Modified++
				// Mark modified files with chimney puff for 1 second
				m.repo.SetFileState(e.Path, domain.StateModified, 1*time.Second)
			case domain.Remove, domain.Rename:
				m.repo.Stats.Deleted++
				// Mark deleted files with demolition for 1 second
				m.repo.SetFileState(e.Path, domain.StateDeleted, 1*time.Second)
			}
		}
		// Rebuild the tree to reflect actual filesystem state
		repo, _ := scan.BuildTree(m.root, m.cfg)
		// Preserve animation states from old repo
		m.preserveAnimationStates(repo)
		m.repo = repo
		return m, waitEvents(m.out)
	}
	return m, nil
}

func (m Model) View() string {
	theme := render.ThemeByName(m.cfg.Theme)
	
	if m.showHelp {
		return render.ViewWithHelp(m.scene, theme, m.width, m.height, m.paused, m.filterActive, m.cfg.Theme)
	}
	
	return render.ViewWithStatus(m.scene, theme, m.width, m.height, m.paused, m.filterActive, m.cfg.Theme)
}

func tick(fps int) tea.Cmd {
	d := time.Second / time.Duration(max(1, fps))
	return func() tea.Msg { time.Sleep(d); return tickMsg(time.Now()) }
}

func waitEvents(ch chan watch.EventOut) tea.Cmd {
	return func() tea.Msg { return eventsMsg(<-ch) }
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// preserveAnimationStates copies active animation states from old repo to new repo
func (m *Model) preserveAnimationStates(newRepo *domain.RepoState) {
	if m.repo == nil {
		return
	}
	for path, oldNode := range m.repo.Index {
		if newNode, exists := newRepo.Index[path]; exists && oldNode.IsStateActive() {
			newNode.State = oldNode.State
			newNode.StateTime = oldNode.StateTime
			newNode.StateExpiry = oldNode.StateExpiry
		}
	}
}
