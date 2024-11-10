.PHONY: all
all:

.PHONY: test
test:
	go test -v -race ./internal/...

.PHONY: generate
generate:
	go generate ./internal/...

.PHONY: lint
lint:
	$(MAKE) -C tools
	./tools/bin/golangci-lint run
