.PHONY: all
all:

.PHONY: test
test:
	go test -v -race ./internal/...

.PHONY: generate
generate:
	$(MAKE) -C tools
	./tools/bin/wire ./internal/di
	# TODO:
	# rm -fr mocks/
	# ./tools/bin/mockery

.PHONY: lint
lint:
	$(MAKE) -C tools
	./tools/bin/golangci-lint run
