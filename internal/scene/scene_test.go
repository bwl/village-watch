package scene

import (
	"testing"

	"example.com/village-watch/internal/domain"
)

func mockRepoDir(name string, x string) *domain.RepoState {
	r := domain.NewRepo("/")
	root := &domain.FileNode{Path: "/", Name: "/", IsDir: true}
	r.Root = root
	r.Upsert(root)
	d := &domain.FileNode{Path: "/" + name, Name: name, IsDir: true}
	r.Upsert(d)
	root.Children = append(root.Children, d)
	return r
}

func TestLabelsToggleAndDraw(t *testing.T) {
	repo := mockRepoDir("foobar", "")
	sc := Derive(repo, 40, 10, true)
	if sc.LabelsVisible {
		t.Fatalf("labels should be false by default")
	}
	// enable and draw labels
	sc.LabelsVisible = true
	sc.DrawLabels(repo)
	// ensure virtual map has at least one rune from name somewhere near top rows
	found := false
	for y := 0; y < VirtualMapHeight && !found; y++ {
		for x := 0; x < VirtualMapWidth; x++ {
			if sc.VirtualMap[y][x] == 'f' || sc.VirtualMap[y][x] == 'F' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected label runes to be drawn when LabelsVisible is true")
	}
}

func TestDeriveWithFPSConsistentCanvasSize(t *testing.T) {
	repo := mockRepoDir("pkg", "")
	sc := DeriveWithFPS(repo, 80, 24, true, 60)
	if len(sc.Canvas) != 24 {
		t.Fatalf("expected canvas height 24, got %d", len(sc.Canvas))
	}
	for i, ln := range sc.Canvas {
		if len([]rune(ln)) != 80 {
			t.Fatalf("line %d width = %d, want 80", i, len([]rune(ln)))
		}
	}
}
