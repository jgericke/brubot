
# Makefile for BruBot

build:
	GOOS=linux GOARCH=amd64 go build cmd/brubot/main.go -o bin/brubot -v

run:
	go run cmd/brubot/main.go