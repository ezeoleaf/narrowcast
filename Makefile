.PHONY: build run clean test lint deps help release all

VERSION ?= 0.3.0
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Build the application
build:
	go build $(LDFLAGS) -o bin/narrowcast .

# Build and run one cycle (mock TTS unless config says otherwise)
run: build
	./bin/narrowcast -once

# Clean build artifacts
clean:
	rm -rf bin/ dist/

lint:
	golangci-lint run

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod tidy

# Release binaries (same layout as .github/workflows/release.yml)
release:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/narrowcast-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/narrowcast-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/narrowcast-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/narrowcast-windows-amd64.exe .

# Show help
help:
	@echo "Available commands:"
	@echo "  build   - Build bin/narrowcast"
	@echo "  run     - Build and run one cycle (-once)"
	@echo "  clean   - Remove bin/ and dist/"
	@echo "  test    - Run tests"
	@echo "  lint    - Run golangci-lint"
	@echo "  deps    - go mod tidy"
	@echo "  release - Build dist/* release binaries (darwin/linux/windows + arm64)"
	@echo "  help    - Show this help message"

# Default target
all: build
