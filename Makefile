
APP_NAME = compress

build:
	go build -o bin/$(APP_NAME) main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(APP_NAME)

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(APP_NAME).exe