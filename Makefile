.PHONY: all build test release clean install

all: build

build:
	go build -o tetris ./cmd/tetris

test:
	go test -v ./...

release:
	./scripts/build-release.sh

install:
	go install ./cmd/tetris

clean:
	rm -rf dist tetris
