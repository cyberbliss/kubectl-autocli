BINARY = kubectl-ag

.PHONY: test
test:
	go test ./... -mod=mod -v

.PHONY: build
build: test
	go build -mod=mod -o releases/darwin/$(BINARY)
