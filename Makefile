build:
	env GOOS=darwin GOARCH=amd64 go build -o bin/alertika.darwin.amd64 main.go
	env GOOS=linux GOARCH=amd64 go build -o bin/alertika.linux.amd64 main.go
	env GOOS=linux GOARCH=arm64 go build -o bin/alertika.linux.arm64 main.go

docker:
	docker build -t alertika .

all: build docker
