.PHONY: build test clean install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o ccq .

test:
	go test ./... -v

clean:
	rm -f ccq

install:
	mkdir -p ~/.local/bin
	go build -ldflags "-X main.Version=$(VERSION)" -o ~/.local/bin/ccq .
