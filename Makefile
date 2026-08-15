MAKEFLAGS += --silent

ldflags := -X 'github.com/yukiteruamano/koma/constant.BuiltAt=$(shell date -u)'
ldflags += -X 'github.com/yukiteruamano/koma/constant.BuiltBy=$(shell whoami)'
ldflags += -X 'github.com/yukiteruamano/koma/constant.Revision=$(shell git rev-parse --short HEAD)'
ldflags += -s
ldflags += -w

build_flags := -ldflags=${ldflags}

all: help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the koma binary"
	@echo "  install      Install the koma binary"
	@echo "  uninstall    Uninstall the koma binary"
	@echo "  test         Run the tests"
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

.PHONY: check-deps-version
check-deps-version:
	@./scripts/check-deps.sh

.PHONY: check-deps-update
check-deps-update:
	@./scripts/check-deps.sh --update

uninstall:
	@rm -f $(shell which koma)

gif:
	@vhs assets/tui.tape
	@vhs assets/inline.tape
