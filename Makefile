# Удобные команды курса. Номер модуля передаётся через M, например: make ci M=03
M ?=
DIR = $(firstword $(wildcard $(M)-*))

.PHONY: help ci ci-all run test cover fmt check-m

help: ## Показать список команд
	@echo "Команды курса:"
	@echo "  make ci M=01      — 🚦 CI-пайплайн модуля (gofmt → build → vet → tests -race → coverage)"
	@echo "  make ci-all       — пайплайн по всем модулям + сводка прогресса"
	@echo "  make run M=01     — запустить модуль (go run)"
	@echo "  make test M=01    — подробный вывод тестов (go test -race -v)"
	@echo "  make cover M=12   — отчёт о покрытии в браузере"
	@echo "  make fmt M=01     — отформатировать код модуля (gofmt -w)"

check-m:
	@if [ -z "$(M)" ] || [ -z "$(DIR)" ]; then echo "Укажи модуль: make <команда> M=01"; exit 2; fi

bin/ci: tools/ci/main.go
	@go build -o bin/ci ./tools/ci

ci: check-m bin/ci
	@./bin/ci $(M)

ci-all: bin/ci
	@./bin/ci all

run: check-m
	go run ./$(DIR)

test: check-m
	go test -race -count=1 -v ./$(DIR)/...

cover: check-m
	go test -coverprofile=cover.out ./$(DIR)/ && go tool cover -html=cover.out

fmt: check-m
	gofmt -w $(DIR)
