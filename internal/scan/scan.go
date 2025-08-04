// internal/scan/scan.go
package scan

import (
	"io/fs"
	"path/filepath"
	"time"

	"example.com/village-watch/internal/config"
	"example.com/village-watch/internal/domain"
)

func BuildTree(root string, cfg config.Config) (*domain.RepoState, error) {
	repo := domain.NewRepo(root)
	rootNode := &domain.FileNode{Path: root, Name: filepath.Base(root), IsDir: true}
	repo.Root = rootNode
	repo.Index[root] = rootNode

	ignore := func(p string) bool {
		for _, g := range cfg.Watch.Ignore {
			if match, _ := filepath.Match(g, filepath.Base(p)); match {
				return true
			}
			if len(g) > 0 && g[len(g)-1] == '/' && stringsHasPathPrefix(p, filepath.Join(root, g[:len(g)-1])) {
				return true
			}
		}
		return false
	}
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == root {
			return nil
		}
		if ignore(path) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return nil
		}
		n := &domain.FileNode{
			Path: path, Name: d.Name(), IsDir: d.IsDir(), ModTime: info.ModTime(), Size: info.Size(), Ext: domain.Ext(d.Name()),
		}
		repo.Upsert(n)
		parent := filepath.Dir(path)
		if par, ok := repo.Index[parent]; ok {
			par.Children = append(par.Children, n)
		} else {
			// ensure parent exists
			par = &domain.FileNode{Path: parent, Name: filepath.Base(parent), IsDir: true}
			repo.Upsert(par)
			par.Children = append(par.Children, n)
		}
		return nil
	})
	repo.LastRefresh = time.Now()
	return repo, nil
}

func stringsHasPathPrefix(p, prefix string) bool {
	pp := filepath.Clean(p)
	pf := filepath.Clean(prefix)
	return len(pp) >= len(pf) && pp[:len(pf)] == pf
}
