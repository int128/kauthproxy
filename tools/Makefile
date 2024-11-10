GOBIN := $(CURDIR)/bin
export GOBIN

all: bin/mockgen bin/wire bin/golangci-lint

bin/mockgen:
	go install go.uber.org/mock/mockgen

bin/wire:
	go install github.com/google/wire/cmd/wire

bin/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint
