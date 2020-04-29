BINARY = kubectl-ag

.PHONY: test
test:
	go test ./... -cover -mod=mod -v

.PHONY: build
build: test
	go build -mod=mod -o releases/darwin/$(BINARY)
