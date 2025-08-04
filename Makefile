# Village Watch Makefile

.PHONY: build test run update-readme help

# Build the village-watch binary
build:
	go build -o village-watch ./cmd/village-watch

# Run tests
test:
	go test ./...

# Run village-watch on current directory
run:
	go run ./cmd/village-watch --path=.

# Test village layout generation
test-layout:
	go run ./cmd/village-watch --test --path=.

# Update README with fresh village layout sample
update-readme:
	./scripts/update-readme-sample.sh

# Clean build artifacts
clean:
	rm -f village-watch

# Install dependencies
deps:
	go mod tidy

# Show help
help:
	@echo "Village Watch - Filesystem Village Visualizer"
	@echo ""
	@echo "Available commands:"
	@echo "  build          Build the village-watch binary"
	@echo "  test           Run Go tests"
	@echo "  run            Run village-watch on current directory"
	@echo "  test-layout    Generate test village layout"
	@echo "  update-readme  Update README with fresh village sample"
	@echo "  clean          Clean build artifacts"
	@echo "  deps           Install dependencies"
	@echo "  help           Show this help message"