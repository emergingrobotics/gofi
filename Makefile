.PHONY: all build test lint clean coverage examples examples-clean examples-test utilities utilities-clean install mac-ping help

# All examples
EXAMPLES := basic crud errors concurrent websocket list fixedips addfixedip delfixedip switches

# All utilities
UTILITIES := gofips gofimac

# Install destination. Defaults to the user's ~/bin so no sudo is needed.
# Override with: make install INSTALL_DIR=/somewhere/else
INSTALL_DIR ?= $(HOME)/bin

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

# Install utilities to $(INSTALL_DIR) (~/bin by default, no sudo needed)
install: utilities
	@mkdir -p $(INSTALL_DIR)
	@for util in $(UTILITIES); do \
		echo "Installing $$util to $(INSTALL_DIR)/$$util"; \
		install -m 755 bin/$$util $(INSTALL_DIR)/$$util; \
	done
	@echo "All utilities installed to $(INSTALL_DIR)."
	@case ":$$PATH:" in *":$(INSTALL_DIR):"*) ;; *) echo "Note: $(INSTALL_DIR) is not on your PATH.";; esac

# === Network probes ===

# ARP-ping a MAC on the local segment. Requires Habets 'arping' (apt install arping);
# only that variant can target a MAC address rather than an IP. Needs root, so it
# invokes sudo. The interface is auto-detected from the default route unless IFACE is set.
# Usage: make mac-ping MAC=aa:bb:cc:dd:ee:ff [IFACE=eth0]
mac-ping:
	@command -v arping >/dev/null 2>&1 || { echo "arping not found (try: sudo apt install arping)"; exit 1; }
	@test -n "$(MAC)" || { echo "Usage: make mac-ping MAC=<mac-address> [IFACE=<interface>]"; exit 1; }
	@iface="$(IFACE)"; \
	if [ -z "$$iface" ]; then iface=$$(ip route show default 2>/dev/null | awk '{print $$5; exit}'); fi; \
	if [ -z "$$iface" ]; then echo "Could not detect a default interface; pass IFACE=<interface>"; exit 1; fi; \
	echo "arping $(MAC) on $$iface..."; \
	sudo arping -c 3 -i "$$iface" "$(MAC)"

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
	@echo "  install         Build and install utilities to ~/bin (override INSTALL_DIR)"
	@echo ""
	@echo "Network probes:"
	@echo "  mac-ping        ARP-ping a MAC on the local segment (MAC=<mac> [IFACE=<if>])"
