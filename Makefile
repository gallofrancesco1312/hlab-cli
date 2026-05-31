BINARY_CLI   := hlab
BINARY_AGENT := hlab-agent
MODULE       := github.com/gallofrancesco1312/hlab-cli

# git describe ritorna il tag più recente (es. v0.1.0-3-gabcdef).
# Se non ci sono tag usa "dev".
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# -trimpath rimuove i percorsi locali dall'eseguibile (riproducibilità).
# -ldflags inietta la variabile version nel binario senza hard-coding.
LDFLAGS      := -trimpath -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build build-cli build-agent run install clean test lint

## build: compila entrambi i binari in ./bin/
build: build-cli build-agent

build-cli:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_CLI) ./cmd/hlab

build-agent:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_AGENT) ./cmd/hlab-agent

## run: esegue la CLI direttamente (utile in sviluppo)
run:
	go run ./cmd/hlab $(ARGS)

## install: installa hlab in $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/hlab

## test: esegue tutti i test
test:
	go test ./...

## lint: esegue il linter (richiede golangci-lint)
lint:
	golangci-lint run ./...

## clean: rimuove i binari compilati
clean:
	rm -rf bin/

## help: mostra questo messaggio
help:
	@grep -E '^##' Makefile | sed 's/## //'
