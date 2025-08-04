// internal/watch/watch.go
package watch

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"example.com/village-watch/internal/config"
	"example.com/village-watch/internal/domain"
)

type EventOut struct {
	Events []domain.FsEvent
}

func Start(root string, cfg config.Config) (chan EventOut, func() error, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	out := make(chan EventOut, 4)
	debounce := time.Duration(cfg.Watch.DebounceMS) * time.Millisecond

	// walk and add watchers for dirs
	addDir := func(path string) error {
		return w.Add(path)
	}
	// initial: watch root + subdirs
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && !ignored(p, cfg) {
			_ = addDir(p)
		}
		return nil
	})
	go func() {
		buf := make([]domain.FsEvent, 0, 64)
		var last time.Time
		flush := func() {
			if len(buf) == 0 {
				return
			}
			out <- EventOut{Events: append([]domain.FsEvent{}, buf...)}
			buf = buf[:0]
		}
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				k := mapKind(ev)
				if k == -1 {
					continue
				}
				if ignored(ev.Name, cfg) {
					continue
				}
				buf = append(buf, domain.FsEvent{Path: ev.Name, Kind: k, When: time.Now()})
				last = time.Now()
			case <-time.After(debounce):
				if time.Since(last) >= debounce {
					flush()
				}
			case err := <-w.Errors:
				_ = err // could log
			}
		}
	}()
	stop := func() error { defer close(out); return w.Close() }
	return out, stop, nil
}

func mapKind(ev fsnotify.Event) domain.EventKind {
	if ev.Op&fsnotify.Create == fsnotify.Create {
		return domain.Create
	}
	if ev.Op&fsnotify.Write == fsnotify.Write {
		return domain.Write
	}
	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		return domain.Remove
	}
	if ev.Op&fsnotify.Rename == fsnotify.Rename {
		return domain.Rename
	}
	return -1
}

func ignored(path string, cfg config.Config) bool {
	base := filepath.Base(path)
	for _, g := range cfg.Watch.Ignore {
		if m, _ := filepath.Match(g, base); m {
			return true
		}
		if strings.HasSuffix(g, "/") && strings.Contains(path, g) {
			return true
		}
	}
	return false
}
