.PHONY: all
all:

.PHONY: test
test:
	go test -v -race ./internal/...

.PHONY: generate
generate:
	go tool github.com/google/wire/cmd/wire ./internal/di
	rm -fr internal/mocks
	go tool go.uber.org/mock/mockgen -destination internal/mocks/mock_browser/mock.go github.com/int128/kauthproxy/internal/browser Interface
	go tool go.uber.org/mock/mockgen -destination internal/mocks/mock_env/mock.go github.com/int128/kauthproxy/internal/env Interface
	go tool go.uber.org/mock/mockgen -destination internal/mocks/mock_portforwarder/mock.go github.com/int128/kauthproxy/internal/portforwarder Interface
	go tool go.uber.org/mock/mockgen -destination internal/mocks/mock_resolver/mock.go github.com/int128/kauthproxy/internal/resolver FactoryInterface,Interface
	go tool go.uber.org/mock/mockgen -destination internal/mocks/mock_reverseproxy/mock.go github.com/int128/kauthproxy/internal/reverseproxy Interface,Instance

.PHONY: lint
lint:
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run
