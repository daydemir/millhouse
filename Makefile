.PHONY: build install clean test

VERSION ?= 0.4.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/daydemir/milhouse/internal/cli.Version=$(VERSION) \
           -X github.com/daydemir/milhouse/internal/cli.GitCommit=$(GIT_COMMIT) \
           -X github.com/daydemir/milhouse/internal/cli.BuildDate=$(BUILD_DATE)

# Build the mill binary
build:
	go build -ldflags "$(LDFLAGS)" -o mil ./cmd/mil

# Install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/mil

# Clean build artifacts
clean:
	rm -f mil

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
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mil-darwin-amd64 ./cmd/mil
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o mil-darwin-arm64 ./cmd/mil
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mil-linux-amd64 ./cmd/mil
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mil-windows-amd64.exe ./cmd/mil

# Release with GoReleaser (dry-run for local testing)
release:
	goreleaser release --snapshot --clean

# Tag and push a release (usage: make tag VERSION=0.1.1)
tag:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)
