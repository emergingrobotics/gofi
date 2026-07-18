.PHONY: all build test lint clean coverage examples examples-clean examples-test utilities utilities-clean install help

# All examples
EXAMPLES := basic crud errors concurrent websocket list fixedips addfixedip delfixedip switches

# All utilities
UTILITIES := gofip gofips gofimac

.DEFAULT_GOAL := help

all: lint test build utilities

build:
	go build ./...

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run ./...

clean: examples-clean utilities-clean
	go clean ./...
	rm -rf coverage.out coverage.html

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# === Examples ===

# Build all examples (always rebuilds)
examples:
	@mkdir -p bin
	@for ex in $(EXAMPLES); do \
		echo "Building $$ex..."; \
		go build -o bin/$$ex ./examples/$$ex; \
	done
	@echo "All examples built in bin/"

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

# === Utilities ===

# Build all utilities
utilities:
	@mkdir -p bin
	@for util in $(UTILITIES); do \
		echo "Building $$util..."; \
		go build -o bin/$$util ./utilities/$$util; \
	done
	@echo "All utilities built in bin/"

# Clean utility binaries
utilities-clean:
	rm -rf bin/

# Install utilities to /usr/local/bin (requires sudo)
install: utilities
	@for util in $(UTILITIES); do \
		echo "Installing $$util to /usr/local/bin/$$util"; \
		sudo install -m 755 bin/$$util /usr/local/bin/$$util; \
	done
	@echo "All utilities installed."

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
	@echo "  examples        Build all examples to bin/"
	@echo "  examples-clean  Remove example binaries"
	@echo "  examples-test   Verify all examples compile"
	@echo ""
	@echo "Utility targets:"
	@echo "  utilities       Build all utilities to bin/"
	@echo "  utilities-clean Remove utility binaries"
	@echo "  install         Build and install utilities to /usr/local/bin (sudo)"
