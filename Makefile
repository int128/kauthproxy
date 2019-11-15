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
    # make the zip files for GitHub Releases
	VERSION=$(CIRCLE_TAG) goxzst -d dist/gh/ -i "LICENSE" -o "$(TARGET)" -t "kauthproxy.rb auth-proxy.yaml" -- -ldflags "$(LDFLAGS)"
	zipinfo dist/gh/kauthproxy_linux_amd64.zip
	# make the Homebrew formula
	mv dist/gh/kauthproxy.rb dist/
	# make the yaml for krew-index
	mkdir -p dist/plugins
	cp dist/gh/auth-proxy.yaml dist/plugins/auth-proxy.yaml

.PHONY: release
release: dist
    # publish to GitHub Releases
	ghr -u "$(CIRCLE_PROJECT_USERNAME)" -r "$(CIRCLE_PROJECT_REPONAME)" "$(CIRCLE_TAG)" dist/gh/
	# publish to Homebrew tap repository
	ghcp commit -u "$(CIRCLE_PROJECT_USERNAME)" -r "homebrew-$(CIRCLE_PROJECT_REPONAME)" -m "$(CIRCLE_TAG)" -C dist/ kauthproxy.rb
	# fork krew-index and create a branch
	ghcp fork-commit -u kubernetes-sigs -r krew-index -b "auth-proxy-$(CIRCLE_TAG)" -m "Bump auth-proxy to $(CIRCLE_TAG)" -C dist/ plugins/auth-proxy.yaml

.PHONY: clean
clean:
	-rm $(TARGET)
	-rm $(TARGET_PLUGIN)
	-rm -r dist/
	-rm coverage.out
