# Village Watch (seed)

A Bubble Tea–based TUI that visualizes a folder as a cozy roguelike village and runs in watch mode on a side monitor.

> Seed project: minimal, runnable skeleton. It scans a directory, watches for changes, and renders a simple village-style grid using Lip Gloss. You can iterate from here (animations, richer tiles, themes).

## Requirements
- Go 1.22+
- A terminal with TrueColor recommended.

## Install & Run
```bash
git init village-watch && cd village-watch
# Copy these files in, or unzip the provided archive.
go mod tidy
go run ./cmd/village-watch --path=. --theme=forest --fps=20
```

## Flags
```
--path=<dir>         Directory to visualize (default: .)
--fps=<n>            Target frames per second (default: 20)
--theme=<name>       forest|seaside|desert|contrast (default: forest)
--no-unicode         Force ASCII-only tiles
--ignore=<comma>     Extra ignore globs (comma-separated)
--test               Generate test village layout and exit
```

## Sample Village Layout

Here's what Village Watch generates for this project:

```
=== Village Layout Test ===
Generated village with 10 key buildings:

░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░▣▬▬▬▬▬▬▣░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▣▬▬▬▬▬▬▣░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░▮⌂····◊▮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▮⌂····◊▮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░▮······▮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▣▬▬▬▬▬▬▣▫▫▫▫▫▫▫▫▫▮······▮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░▮······▮▫▫▫▫░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▮⌂····◊▮░░░░░░░░░▮······▮░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
░░░░░░░░░░░▮······▮░░░▫░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▮······▮░░░░░░░░░▣▬▬▬▫▬▬▣░░░░░░░░▣▬▬▬▬▬▬▣░░░░░░░░▣▬▬▬▬▬▬▣░░░░░░░░░░░░░░░
░░░░░░░░░░░▣▬▬▬▫▬▬▣░░░▫░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▮······▮░░░░░░░░░░░░░▫░░░░░░░░░▮⌂····◊▮░░░░░░░░▮⌂····◊▮░░░░░░░░░░░░░░░
░░░░░░░░░░░░░░░░░░░░░░▫░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▮······▮░░░░░░░░░░░░░▫░░░░░░░░░▮······▮░░░░░░░░▮······▮░░░░░░░░░░░░░░░

Legend: ▣ = house frame, ▮ = walls, ⌂ = roof, ◊ = door, · = interior, ▫ = roads, ░ = grass
```

Test your own layout: `go run ./cmd/village-watch --test --path=.`

## Config (optional `village.yml`)
```yaml
theme: forest
fps: 20
watch:
  debounce_ms: 200
  ignore:
    - ".git/"
    - "node_modules/"
mapping:
  ".md": library
  ".yaml": kiosk
render:
  unicode: true
  lod_thresholds: { level1: 400, level2: 1200 }
```
Run with config: `go run ./cmd/village-watch --path=.`, it will load `village.yml` if present.

## Roadmap (you can extend)
- Add Harmonica for eased build/demolition animations.
- Git banners (untracked/modified/staged).
- Mini-map and pan/zoom.
- Log-driven lantern brightness with a tailer.
