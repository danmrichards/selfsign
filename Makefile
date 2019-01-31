GOARCH=amd64
GO111MODULE=on

build: linux darwin windows

linux: client-linux server-http-linux server-https-linux

darwin: client-darwin server-http-darwin server-https-darwin

windows: client-windows server-http-windows server-https-windows

client-linux:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=linux go build -o ./bin/client-linux-${GOARCH} ./client

client-darwin:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=darwin go build -o ./bin/client-darwin-${GOARCH} ./client

client-windows:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=windows go build -o ./bin/client-windows-${GOARCH}.exe ./client

server-http-linux:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=linux go build -o ./bin/server-http-linux-${GOARCH} ./server/http

server-http-darwin:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=darwin go build -o ./bin/server-http-darwin-${GOARCH} ./server/http

server-http-windows:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=windows go build -o ./bin/server-http-windows-${GOARCH}.exe ./server/http

server-https-linux:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=linux go build -o ./bin/server-https-linux-${GOARCH} ./server/https

server-https-darwin:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=darwin go build -o ./bin/server-https-darwin-${GOARCH} ./server/https

server-https-windows:
		CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=windows go build -o ./bin/server-https-windows-${GOARCH}.exe ./server/https

.PHONY: build
