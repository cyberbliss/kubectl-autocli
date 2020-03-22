test:
	go test ./... -mod=mod -v

linux: test
	mkdir -p ./releases/linux
	GOARCH=amd64 GOOS=linux go build -mod=mod
	mv autocli ./releases/linux

osx: test
	mkdir -p ./releases/darwin
	GOARCH=amd64 GOOS=darwin go build -mod=mod
	mv autocli ./releases/darwin/amd64