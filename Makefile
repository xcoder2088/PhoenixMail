.PHONY: run test build
run:
	go run ./cmd/phoenixmail

test:
	go test ./...

build:
	go build -o bin/phoenixmail ./cmd/phoenixmail
