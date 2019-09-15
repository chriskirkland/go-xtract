GO111MODULE := on
export
LINT_VERSION="1.17.1"


.PHONY: lintall
lintall: fmt lint

.PHONY: install
install:
	go install github.com/chriskirkland/go-xtract/cmd/xtract

.PHONY: fmt
fmt: deps
	golangci-lint run --disable-all --enable=gofmt --fix

.PHONY: dofmt
dofmt: deps
	golangci-lint run --disable-all --enable=gofmt --fix

.PHONY: deps
deps:
	@if ! which golangci-lint >/dev/null || [[ "$$(golangci-lint --version)" != *${LINT_VERSION}* ]]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.17.1; \
	fi

.PHONY: lint
lint: deps
	golangci-lint run

.PHONY: test
test: i18n
	go test -race -covermode=atomic -coverprofile=cover.out ./...

.PHONY: integration
integration: install
	cd _integration && go test ./... -v

.PHONY: coverage
coverage: test
	go tool cover -html=cover.out -o=cover.html
