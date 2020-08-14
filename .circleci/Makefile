.PHONY: all
all:

.PHONY: install-deps
install-deps:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(HOME)/go/bin v1.30.0
	go get -v github.com/int128/goxzst github.com/int128/ghcp
