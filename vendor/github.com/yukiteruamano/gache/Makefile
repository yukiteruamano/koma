.PHONY: test lint vet cover clean bench

GO ?= go
GOLANGCI_LINT ?= golangci-lint

test:
	$(GO) test -race -count=1 -v ./...

vet:
	$(GO) vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

cover:
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

bench:
	$(GO) test -bench=. -benchmem ./...

clean:
	rm -f coverage.out coverage.html
	rm -rf test/test.json test/test.json.tmp testdata/tmp

fmt:
	$(GO) fmt ./...
