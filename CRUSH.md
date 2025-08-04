# CRUSH.md

Project: village-watch (Go 1.22)

Build/run
- make build            # go build -o village-watch ./cmd/village-watch
- make run              # go run ./cmd/village-watch --path=.
- make test             # go test ./...
- Single test file      # go test ./internal/packagename -run TestName
- Verbose/specific      # go test ./... -run 'Regex' -v
- Update sample         # make update-readme
- Tidy deps             # make deps
- Clean                 # make clean

Lint/format
- gofmt -w .            # format code
- go vet ./...          # static checks
- golangci-lint run     # if repo has it installed; otherwise skip

Testing notes
- Use table-driven tests and t.Helper; keep tests in *_test.go next to code
- Prefer deterministic tests; avoid filesystem side effects or use t.TempDir

Code style
- Imports: stdlib first, then external, then module-local; grouped and gofmt-sorted
- Naming: Exported identifiers use CamelCase with clear nouns; unexported use lowerCamelCase
- Errors: return errors, wrap with context using fmt.Errorf("...: %w", err); no panics in libs
- Types: prefer concrete types at boundaries; use interfaces in consumers when needed
- Context: pass context.Context for cancellable/IO ops; don’t store in structs
- Logging/IO: avoid fmt.Print in library code; return data and errors instead
- Packages: keep small, cohesive packages under internal/*; avoid cyclic deps
- Rendering/UI: follow existing Bubble Tea patterns; keep state in model, pure render in view
- Config: keep defaults in code; load YAML via internal/config; validate on load

Contributing
- Run format/vet/test before PRs; keep functions small; avoid global vars; document exported APIs
