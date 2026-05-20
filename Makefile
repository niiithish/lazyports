.PHONY: build install test run clean setup-go

BINARY := lazyports
GO := ./scripts/go
GOBIN ?= $(HOME)/.local/bin

build:
	$(GO) build -o $(BINARY) .

install: build
	./install.sh

test:
	$(GO) test ./... -v

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

setup-go:
	./scripts/setup-go.sh
