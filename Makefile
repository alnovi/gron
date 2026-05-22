.SILENT:
.DEFAULT_GOAL := help

## lint: статический анализ кода
.PHONY: lint
lint:
	@go tool golangci-lint run ./...

## lint-fix: статический анализ (авто-исправление)
.PHONY: lint-fix
lint-fix:
	go tool golangci-lint run ./... --fix --timeout 650s

## test: запуск тестов
.PHONY: test
test:
	@go test -count=1 -coverpkg=./... -coverprofile=./coverage ./...
	@go tool cover -html=./coverage
	@rm ./coverage

## help: справка
.PHONY: help
help:
	@echo 'Gron'
	@echo ''
	@echo 'Usage:'
	@echo '  make <command>'
	@echo ''
	@echo 'The commands are:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'