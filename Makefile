TARGET := kauthproxy
TARGET_PLUGIN := kubectl-auth_proxy
CIRCLE_TAG ?= HEAD
LDFLAGS := -X main.version=$(CIRCLE_TAG)

.PHONY: all
all: $(TARGET)

.PHONY: check
check:
	golangci-lint run
	go test -v -race -cover -coverprofile=coverage.out ./...

$(TARGET): $(wildcard *.go)
	go build -o $@ -ldflags "$(LDFLAGS)"

$(TARGET_PLUGIN): $(TARGET)
	ln -sf $(TARGET) $@

.PHONY: run
run: $(TARGET_PLUGIN)
	PATH=.:$(PATH) kubectl auth-proxy --help

dist:
	VERSION=$(CIRCLE_TAG) goxzst -d dist/gh/ -o "$(TARGET)" -t "kauthproxy.rb auth-proxy.yaml" -- -ldflags "$(LDFLAGS)"
	mv dist/gh/kauthproxy.rb dist/
	mkdir -p dist/plugins
	cp dist/gh/auth-proxy.yaml dist/plugins/auth-proxy.yaml

.PHONY: release
release: dist
	ghr -u "$(CIRCLE_PROJECT_USERNAME)" -r "$(CIRCLE_PROJECT_REPONAME)" "$(CIRCLE_TAG)" dist/gh/
	ghcp commit -u "$(CIRCLE_PROJECT_USERNAME)" -r "homebrew-$(CIRCLE_PROJECT_REPONAME)" -m "$(CIRCLE_TAG)" -C dist/ kauthproxy.rb
	ghcp fork-commit -u kubernetes-sigs -r krew-index -b "auth-proxy-$(CIRCLE_TAG)" -m "Bump auth-proxy to $(CIRCLE_TAG)" -C dist/ plugins/auth-proxy.yaml

.PHONY: clean
clean:
	-rm $(TARGET)
	-rm $(TARGET_PLUGIN)
	-rm -r dist/
	-rm coverage.out
