# Local development for the tuikit rebuilds.
#
# `make check` is what you run before pushing. Every rebuild's screens are held
# to a golden at two terminal sizes, and gcpeasy's guards run here too.

.PHONY: help check test lint shots run

## help: this
help:
	@grep -h '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

## test: every rebuild's screens against its goldens
test:
	go test ./...

# Pinned to the same version tuikit runs, and through `go run` so it is the
# same whether or not anything is installed.
GOLANGCI ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0

## lint: gofmt, vet, golangci-lint
lint:
	gofmt -l . | tee /dev/stderr | (! read)
	go vet ./...
	$(GOLANGCI) run ./...

## check: everything, before you push
check: test lint
	@echo "all checks passed"

TOOLS = lazygit gcpeasy k9s bottom gitui termshark yazi dive fx gh-dash

## shots: regenerate every rebuild's screenshots and screens.md
shots:
	@for t in $(TOOLS); do go run ./$$t -shots; done

## run: open one rebuild — make run T=lazygit
run:
	@go run ./$(T)
