TARGET := kauthproxy
VERSION ?= latest
LDFLAGS := -X main.version=$(VERSION)

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
	VERSION=$(VERSION) CGO_ENABLED=0 goxzst -d dist/output/ -i "LICENSE" -o "$(TARGET)" -t "dist/kauthproxy.rb dist/auth-proxy.yaml" -parallelism 3 -- -ldflags "$(LDFLAGS)"
	# test the zip file
	zipinfo dist/output/kauthproxy_linux_amd64.zip
	# make the krew yaml structure
	mkdir -p dist/output/plugins
	mv dist/output/auth-proxy.yaml dist/output/plugins/auth-proxy.yaml

.PHONY: release
release:
	ghcp release -r "$(GITHUB_REPOSITORY)" -t "$(VERSION)" dist/output/

.PHONY: release-external
release-external:
	# homebrew
	ghcp commit -r int128/homebrew-kauthproxy -b "bump-$(VERSION)" -m "Bump the version to $(VERSION)" -C dist/output/ kauthproxy.rb
	ghcp pull-request -r int128/homebrew-kauthproxy -b "bump-$(VERSION)" --title "Bump the version to $(VERSION)"
	# krew
	ghcp fork-commit -r kubernetes-sigs/krew-index -b "auth-proxy-$(VERSION)" -m "Bump auth-proxy to $(VERSION)" -C dist/output/ plugins/auth-proxy.yaml
	ghcp pull-request -r kubernetes-sigs/krew-index --base-repo "int128/krew-index" -b "auth-proxy-$(VERSION)" --title "Bump auth-proxy to $(VERSION)" --draft

.PHONY: clean
clean:
	-rm $(TARGET)
	-rm -r dist/output/
	-rm coverage.out
