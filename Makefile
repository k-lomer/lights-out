.PHONY: all build run fmt vet lint test clean
all: fmt vet lint test build
build:
	go build -o bin/lights-out ./cmd
run:
	go run ./cmd
fmt:
	go fmt ./...
vet:
	go vet ./...
lint:
	golangci-lint run
test:
	go clean -testcache
	go test ./...
coverage:
	go test -cover ./...
clean:
	rm -rf bin/
