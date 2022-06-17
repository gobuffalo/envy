TAGS ?= ""
GO_BIN ?= go
GO_ENV ?= test

install:
	$(GO_BIN) install -tags ${TAGS} -v .

tidy:
	$(GO_BIN) mod tidy

build:
	$(GO_BIN) build -v .

test:
	$(GO_BIN) test -v -race -cover -tags $(TAGS) ./...

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --enable-all

update:
	rm go.*
	$(GO_BIN) mod init github.com/gobuffalo/envy/v2
	$(GO_BIN) mod tidy
	make test
