.PHONY: all
all:

.PHONY: test
test:
	go test -v -race ./internal/...

.PHONY: generate
generate:
	$(MAKE) -C tools
	./tools/bin/wire ./internal/di
	./tools/bin/mockgen -destination internal/portforwarder/mock_portforwarder/mock_portforwarder.go github.com/int128/kauthproxy/internal/portforwarder Interface
	./tools/bin/mockgen -destination internal/reverseproxy/mock_reverseproxy/mock_reverseproxy.go github.com/int128/kauthproxy/internal/reverseproxy Interface,Instance
	./tools/bin/mockgen -destination internal/browser/mock_browser/mock_browser.go github.com/int128/kauthproxy/internal/browser Interface
	./tools/bin/mockgen -destination internal/resolver/mock_resolver/mock_resolver.go github.com/int128/kauthproxy/internal/resolver FactoryInterface,Interface
	./tools/bin/mockgen -destination internal/env/mock_env/mock_env.go github.com/int128/kauthproxy/internal/env Interface

.PHONY: lint
lint:
	$(MAKE) -C tools
	./tools/bin/golangci-lint run
