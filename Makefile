.PHONY: help build run migrate test clean

help: ## Показать помощь
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Собрать бинарник
	go build -o ./bin/api .

run: ## Запустить сервер
	go run main.go serve

migrate: ## Выполнить миграции
	go run main.go migrate

test: ## Запустить тесты
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Показать покрытие тестами
	go tool cover -html=coverage.out

lint: ## Проверить код линтером
	golangci-lint run

fmt: ## Форматировать код
	go fmt ./...
	gofumpt -l -w .

tidy: ## Обновить зависимости
	go mod tidy

clean: ## Очистить сгенерированные файлы
	rm -rf ./bin
	rm -f coverage.out

docker-build: ## Собрать Docker образ
	docker build -t senderscore-api:latest .

docker-run: ## Запустить в Docker
	docker-compose up -d

docker-stop: ## Остановить Docker
	docker-compose down

dev: ## Запуск в dev режиме с hot reload (требует air)
	air -c air.api.toml

install-tools: ## Установить инструменты разработки
	go install github.com/cosmtrek/air@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.DEFAULT_GOAL := help