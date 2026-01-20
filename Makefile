.PHONY: all build test lint clean coverage examples

all: lint test build

build:
	go build ./...

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	go clean ./...
	rm -rf bin/

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

examples: bin/examples/list

bin/examples/list: examples/list/main.go
	@mkdir -p bin/examples
	go build -o bin/examples/list ./examples/list
