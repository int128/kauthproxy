TARGET := kauthproxy
CIRCLE_TAG ?= latest
LDFLAGS := -X main.version=$(CIRCLE_TAG)

.PHONY: all
all: $(TARGET)

.PHONY: check
check:
	golangci-lint run
	go test -v -race -cover -coverprofile=coverage.out ./...

$(TARGET): $(wildcard **/*.go)
	go build -o $@ -ldflags "$(LDFLAGS)"

.PHONY: dist
dist: dist/output
dist/output:
	# make the zip files for GitHub Releases
	VERSION=$(CIRCLE_TAG) CGO_ENABLED=0 goxzst -d dist/output/ -i "LICENSE" -o "$(TARGET)" -t "dist/kauthproxy.rb dist/auth-proxy.yaml" -- -ldflags "$(LDFLAGS)"
	# test the zip file
	zipinfo dist/output/kauthproxy_linux_amd64.zip
	# make the krew yaml structure
	mkdir -p dist/output/plugins
	mv dist/output/auth-proxy.yaml dist/output/plugins/auth-proxy.yaml

.PHONY: release
release: dist
	# publish to GitHub Releases
	ghcp release -u "$(CIRCLE_PROJECT_USERNAME)" -r "$(CIRCLE_PROJECT_REPONAME)" -t "$(CIRCLE_TAG)" dist/output/
	# publish to Homebrew tap repository
	ghcp commit -u "$(CIRCLE_PROJECT_USERNAME)" -r "homebrew-$(CIRCLE_PROJECT_REPONAME)" -b "bump-$(CIRCLE_TAG)" -m "Bump version to $(CIRCLE_TAG)" -C dist/ kauthproxy.rb
	# create a pull request
	ghcp pull-request -u "$(CIRCLE_PROJECT_USERNAME)" -r "homebrew-$(CIRCLE_PROJECT_REPONAME)" -b "bump-$(CIRCLE_TAG)" --title "Bump version to $(CIRCLE_TAG)"
	# fork krew-index and create a branch
	ghcp fork-commit -u kubernetes-sigs -r krew-index -b "auth-proxy-$(CIRCLE_TAG)" -m "Bump auth-proxy to $(CIRCLE_TAG)" -C dist/output/ plugins/auth-proxy.yaml

.PHONY: clean
clean:
	-rm $(TARGET)
	-rm -r dist/output/
	-rm coverage.out
