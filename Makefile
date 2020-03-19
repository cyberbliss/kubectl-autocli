test:
	go test ./... -mod=mod -v

linux: test
	GOARCH=amd64 GOOS=linux go build -mod=mod
	mv autocli ./releases/linux/amd64

osx: test
	GOARCH=amd64 GOOS=darwin go build -mod=mod
	mv autocli ./releases/darwin/amd64