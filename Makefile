.PHONY: all build run fmt vet lint test clean
all: fmt vet lint test build
build:
	go build -o bin/lights-out .
run:
	go run .
fmt:
	go fmt ./...
vet:
	go vet ./...
lint:
	golangci-lint run
test:
	go test ./...
clean:
	rm -rf bin/