.PHONY: build test clean install

build:
	go build -o ccq .

test:
	go test ./... -v

clean:
	rm -f ccq

install: build
	mkdir -p ~/.local/bin
	cp ccq ~/.local/bin/ccq
