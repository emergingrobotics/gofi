.PHONY: all build test lint clean coverage \
        examples examples-build examples-clean examples-test

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

# Build all examples
examples: examples-build

examples-build: $(addprefix bin/examples/,$(EXAMPLES))

# Pattern rule for building examples
bin/examples/%: examples/%/main.go
	@mkdir -p bin/examples
	go build -o $@ ./examples/$*

# Clean example binaries
examples-clean:
	rm -rf bin/examples

# Test that all examples compile
examples-test:
	@echo "Testing examples compile..."
	@for ex in $(EXAMPLES); do \
		echo "  Checking $$ex..."; \
		go build -o /dev/null ./examples/$$ex || exit 1; \
	done
	@echo "All examples compile successfully."

# Individual example targets
.PHONY: example-basic example-crud example-errors example-concurrent example-websocket example-list

example-basic: bin/examples/basic
example-crud: bin/examples/crud
example-errors: bin/examples/errors
example-concurrent: bin/examples/concurrent
example-websocket: bin/examples/websocket
example-list: bin/examples/list

# Help target
.PHONY: help
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
	@echo "  examples      Build all examples to bin/examples/"
	@echo "  examples-clean  Remove example binaries"
	@echo "  examples-test   Verify all examples compile"
	@echo ""
	@echo "Individual examples:"
	@echo "  example-basic      Build basic example"
	@echo "  example-crud       Build crud example"
	@echo "  example-errors     Build errors example"
	@echo "  example-concurrent Build concurrent example"
	@echo "  example-websocket  Build websocket example"
	@echo "  example-list       Build list example"
