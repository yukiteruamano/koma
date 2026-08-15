MAKEFLAGS += --silent

ldflags := -X 'github.com/yukiteruamano/koma/constant.BuiltAt=$(shell date -u)'
ldflags += -X 'github.com/yukiteruamano/koma/constant.BuiltBy=$(shell whoami)'
ldflags += -X 'github.com/yukiteruamano/koma/constant.Revision=$(shell git rev-parse --short HEAD)'
ldflags += -s
ldflags += -w

build_flags := -ldflags=${ldflags}

all: help

.PHONY: all help build install test clean uninstall gif check-deps-version check-deps-update

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the koma binary"
	@echo "  install      Install the koma binary"
	@echo "  uninstall    Uninstall the koma binary"
	@echo "  test         Run the tests"
	@echo "  clean        Remove build artifacts and Go build/test caches"
	@echo "  check-deps-version   Check Go toolchain, outdated deps and vulnerabilities (advisory)"
	@echo "  check-deps-update    Bump outdated direct dependencies to latest"
	@echo "  gif          Generate usage gifs"
	@echo "  help         Show this help message"
	@echo ""

install:
	@go install "$(build_flags)"


build:
	@go build "$(build_flags)"

test:
	@go test ./...

clean:
	@rm -f koma koma_test
	@rm -rf dist
	@go clean -cache -testcache

.PHONY: check-deps-version
check-deps-version:
	@./scripts/check-deps.sh

.PHONY: check-deps-update
check-deps-update:
	@./scripts/check-deps.sh --update

uninstall:
	@if command -v koma >/dev/null 2>&1; then rm -f $$(command -v koma); else echo "koma is not installed"; fi

gif:
	@command -v vhs >/dev/null 2>&1 || { echo "vhs is required (https://github.com/charmbracelet/vhs)"; exit 1; }
	@vhs assets/tui.tape
	@vhs assets/inline.tape
