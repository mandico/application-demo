NAME=go-producer
VERSION=0.0.1
AUTHOR=mandico

.PHONY: docker-build
## docker-build: Build Image
docker-build:
	@docker build . -t go-producer:0.1.0 --no-cache

.PHONY: run-producer
## run-producer: Run Producer
run-producer:
	@go run main.go localhost:9092 important

.PHONY: deps-install
## deps-install: Install packages and dependencies
deps-install:
	@go mod download

.PHONY: help
all: help
# help: show this help message
help: Makefile
	@echo
	@echo " Choose a command to run in "$(APP_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo