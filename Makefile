.PHONY: all build build-web build-go test clean run

BINARY=bin/formpilot

all: build

build-web:
	cd web && npm install && npm run build

build-go:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/formpilot

build: build-web build-go

test:
	go test -v -race ./...

clean:
	rm -rf bin/ web/dist formpilot.db

run: build-go
	./$(BINARY) serve --config examples/formpilot.yaml
