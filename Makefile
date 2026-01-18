.PHONY: build install clean test fmt lint

build:
	go build -o mil ./cmd/mil

install:
	go install ./cmd/mil

clean:
	rm -f mil

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...
