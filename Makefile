GO ?= go
NPM ?= npm

.PHONY: test test-go test-admin test-prototype test-integration test-all coverage

test: test-go test-admin test-prototype

test-go:
	$(GO) test -race ./src/... ./tests/...

test-admin:
	$(NPM) --prefix assets test

# bin is an optional, independently tracked local repository.
test-prototype:
	@if [ -f bin/proto/package.json ]; then $(NPM) --prefix bin/proto test; else echo 'Local prototype unavailable; skipping.'; fi

test-integration:
	GO='$(GO)' sh scripts/test-integration.sh -race

test-all: test test-integration

coverage:
	mkdir -p coverage
	GO='$(GO)' sh scripts/test-integration.sh -race -coverpkg=./src/... -coverprofile=coverage/backend.out
	$(GO) tool cover -func=coverage/backend.out
	$(NPM) --prefix assets run test:coverage
	@if [ -f bin/proto/package.json ]; then $(NPM) --prefix bin/proto run test:coverage; fi
