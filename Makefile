BINARY 	= kubectl-ac
GO		= go
MODULE	= $(shell $(GO) list -m)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo devx)
DATE    ?= $(shell date +%FT%T%z)

.PHONY: test
test:
	go test ./... -cover -mod=mod -v

.PHONY: build
build: test
	go build -mod=mod -ldflags '-X $(MODULE)/cmd.BuildVersion=$(VERSION) -X $(MODULE)/cmd.BuildTime=$(DATE)' -o releases/darwin/$(BINARY)

.PHONY: fmt
fmt:
	go fmt ./...

lint:
	golint ./...
