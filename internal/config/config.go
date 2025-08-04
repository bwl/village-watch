// internal/config/config.go
package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type WatchCfg struct {
	DebounceMS int      `yaml:"debounce_ms"`
	Ignore     []string `yaml:"ignore"`
}

type RenderCfg struct {
	Unicode      bool           `yaml:"unicode"`
	LODThreshold map[string]int `yaml:"lod_thresholds"`
}

type MappingCfg map[string]string

type Config struct {
	Theme   string     `yaml:"theme"`
	FPS     int        `yaml:"fps"`
	Watch   WatchCfg   `yaml:"watch"`
	Mapping MappingCfg `yaml:"mapping"`
	Render  RenderCfg  `yaml:"render"`
}

func Default() Config {
	return Config{
		Theme:   "forest",
		FPS:     20,
		Watch:   WatchCfg{DebounceMS: 200, Ignore: []string{".git/", "node_modules/", "dist/"}},
		Mapping: MappingCfg{},
		Render:  RenderCfg{Unicode: true, LODThreshold: map[string]int{"level1": 400, "level2": 1200}},
	}
}

func Load(root string) (Config, error) {
	cfg := Default()
	b, err := os.ReadFile(filepath.Join(root, "village.yml"))
	if err != nil {
		return cfg, nil
	}
	_ = yaml.Unmarshal(b, &cfg)
	return cfg, nil
}

func (c *Config) ApplyIgnoreCSV(csv string) {
	if strings.TrimSpace(csv) == "" {
		return
	}
	parts := strings.Split(csv, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			c.Watch.Ignore = append(c.Watch.Ignore, p)
		}
	}
}
