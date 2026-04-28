.PHONY: build run clean

build:
	go build -o bin/lights-out .

run:
	go run .

clean:
	rm -rf bin/