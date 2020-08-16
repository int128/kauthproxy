# CircleCI specific variables
CIRCLE_TAG ?= latest
GITHUB_USERNAME := $(CIRCLE_PROJECT_USERNAME)
GITHUB_REPONAME := $(CIRCLE_PROJECT_REPONAME)

TARGET := kauthproxy
VERSION ?= $(CIRCLE_TAG)
LDFLAGS := -X main.version=$(CIRCLE_TAG)

.PHONY: all
all: $(TARGET)

.PHONY: check
check:
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
	# publish the binaries
	ghcp release -u "$(GITHUB_USERNAME)" -r "$(GITHUB_REPONAME)" -t "$(VERSION)" dist/output/
	# publish the Homebrew formula
	ghcp commit -u "$(GITHUB_USERNAME)" -r "homebrew-$(GITHUB_REPONAME)" -b "bump-$(VERSION)" -m "Bump the version to $(VERSION)" -C dist/output/ kauthproxy.rb
	ghcp pull-request -u "$(GITHUB_USERNAME)" -r "homebrew-$(GITHUB_REPONAME)" -b "bump-$(VERSION)" --title "Bump the version to $(VERSION)"
	# publish the Krew manifest
	ghcp fork-commit -u kubernetes-sigs -r krew-index -b "auth-proxy-$(VERSION)" -m "Bump auth-proxy to $(VERSION)" -C dist/output/ plugins/auth-proxy.yaml

.PHONY: clean
clean:
	-rm $(TARGET)
	-rm -r dist/output/
	-rm coverage.out
