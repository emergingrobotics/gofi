.PHONY: all build test lint clean coverage examples examples-clean examples-test help

# All examples
EXAMPLES := basic crud errors concurrent websocket list

all: lint test build

build:
	go build ./...

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run ./...

clean: examples-clean
	go clean ./...
	rm -rf coverage.out coverage.html

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# === Examples ===

# Build all examples (always rebuilds)
examples:
	@mkdir -p bin/examples
	@for ex in $(EXAMPLES); do \
		echo "Building $$ex..."; \
		go build -o bin/examples/$$ex ./examples/$$ex; \
	done
	@echo "All examples built in bin/examples/"

# Clean example binaries
examples-clean:
	rm -rf bin/

# Test that all examples compile
examples-test:
	@echo "Testing examples compile..."
	@for ex in $(EXAMPLES); do \
		echo "  Checking $$ex..."; \
		go build -o /dev/null ./examples/$$ex || exit 1; \
	done
	@echo "All examples compile successfully."

# Help target
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Main targets:"
	@echo "  all           Run lint, test, and build"
	@echo "  build         Build the module"
	@echo "  test          Run all tests"
	@echo "  lint          Run linter"
	@echo "  clean         Clean all build artifacts"
	@echo "  coverage      Generate coverage report"
	@echo ""
	@echo "Example targets:"
	@echo "  examples        Build all examples to bin/examples/"
	@echo "  examples-clean  Remove example binaries"
	@echo "  examples-test   Verify all examples compile"
