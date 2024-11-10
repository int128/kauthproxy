.PHONY: all
all:

.PHONY: test
test:
	go test -v -race ./internal/...

.PHONY: generate
generate:
	$(MAKE) -C tools
	./tools/bin/wire ./internal/di
	rm -fr internal/mocks
	./tools/bin/mockgen -destination internal/mocks/mock_browser/mock.go github.com/int128/kauthproxy/internal/browser Interface
	./tools/bin/mockgen -destination internal/mocks/mock_env/mock.go github.com/int128/kauthproxy/internal/env Interface
	./tools/bin/mockgen -destination internal/mocks/mock_portforwarder/mock.go github.com/int128/kauthproxy/internal/portforwarder Interface
	./tools/bin/mockgen -destination internal/mocks/mock_resolver/mock.go github.com/int128/kauthproxy/internal/resolver FactoryInterface,Interface
	./tools/bin/mockgen -destination internal/mocks/mock_reverseproxy/mock.go github.com/int128/kauthproxy/internal/reverseproxy Interface,Instance

.PHONY: lint
lint:
	$(MAKE) -C tools
	./tools/bin/golangci-lint run
