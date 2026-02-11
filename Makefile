.PHONY: help build run migrate test clean

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build:
	go build -o ./bin/api .

run:
	go run main.go serve

migrate:
	go run main.go migrate

update:
	go run main.go update

test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test
	go tool cover -html=coverage.out

lint:
	golangci-lint run

fmt:
	go fmt ./...
	gofumpt -l -w .

tidy:
	go mod tidy

clean:
	rm -rf ./bin
	rm -f coverage.out

docker-build:
	docker build -t senderscore-api:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

dev:
	air -c air.api.toml

install-tools:
	go install github.com/cosmtrek/air@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.DEFAULT_GOAL := help