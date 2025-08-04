// internal/domain/domain.go
package domain

import (
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type EventKind int

const (
	Create EventKind = iota
	Write
	Remove
	Rename
)

type FsEvent struct {
	Path string
	Kind EventKind
	When time.Time
}

type FileState int

const (
	StateNormal FileState = iota
	StateNew                     // Recently created, show construction animation
	StateModified                // Recently modified, show chimney puff/glow
	StateDeleted                 // Being deleted, show demolition
)

type FileNode struct {
	Path        string
	Name        string
	Ext         string
	Size        int64
	ModTime     time.Time
	IsDir       bool
	Children    []*FileNode
	State       FileState   // Current animation state
	StateTime   time.Time   // When state was set
	StateExpiry time.Time   // When state expires back to normal
}

type RepoState struct {
	RootPath    string
	Root        *FileNode
	Index       map[string]*FileNode
	Stats       ActivityStats
	LastRefresh time.Time
}

type ActivityStats struct {
	NewFiles int
	Modified int
	Deleted  int
}

func NewRepo(root string) *RepoState {
	return &RepoState{RootPath: root, Index: make(map[string]*FileNode)}
}

func (r *RepoState) Upsert(node *FileNode) {
	r.Index[node.Path] = node
}

func (r *RepoState) Delete(path string) {
	if node, exists := r.Index[path]; exists {
		// Mark as deleted with animation
		node.SetState(StateDeleted, 1*time.Second)
	}
	delete(r.Index, path)
}

// SetFileState marks a file with an animation state for a duration
func (r *RepoState) SetFileState(path string, state FileState, duration time.Duration) {
	if node, exists := r.Index[path]; exists {
		node.SetState(state, duration)
	}
}

// UpdateStates expires old animation states back to normal
func (r *RepoState) UpdateStates() {
	now := time.Now()
	for _, node := range r.Index {
		if node.State != StateNormal && now.After(node.StateExpiry) {
			node.State = StateNormal
		}
	}
}

// SetState on a FileNode with expiry
func (f *FileNode) SetState(state FileState, duration time.Duration) {
	f.State = state
	f.StateTime = time.Now()
	f.StateExpiry = f.StateTime.Add(duration)
}

// IsStateActive checks if animation state is still active
func (f *FileNode) IsStateActive() bool {
	return f.State != StateNormal && time.Now().Before(f.StateExpiry)
}

func (r *RepoState) SortedChildren(dir *FileNode) []*FileNode {
	chs := append([]*FileNode{}, dir.Children...)
	sort.Slice(chs, func(i, j int) bool { return strings.ToLower(chs[i].Name) < strings.ToLower(chs[j].Name) })
	return chs
}

func Ext(name string) string { return strings.ToLower(filepath.Ext(name)) }
