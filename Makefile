.PHONY: build install clean test

VERSION ?= 0.1.9
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/suelio/millhouse/internal/cli.Version=$(VERSION) \
           -X github.com/suelio/millhouse/internal/cli.GitCommit=$(GIT_COMMIT) \
           -X github.com/suelio/millhouse/internal/cli.BuildDate=$(BUILD_DATE)

# Build the mill binary
build:
	go build -ldflags "$(LDFLAGS)" -o mill ./cmd/mill

# Install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/mill

# Clean build artifacts
clean:
	rm -f mill

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Build for all platforms (manual)
release-manual:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mill-darwin-amd64 ./cmd/mill
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o mill-darwin-arm64 ./cmd/mill
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mill-linux-amd64 ./cmd/mill
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mill-windows-amd64.exe ./cmd/mill

# Release with GoReleaser (dry-run for local testing)
release:
	goreleaser release --snapshot --clean

# Tag and push a release (usage: make tag VERSION=0.1.1)
tag:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)
