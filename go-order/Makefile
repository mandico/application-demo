NAME=go-order
VERSION=0.0.1
AUTHOR=mandico

.PHONY: run-producer
## run-producer: Run consumer
run-producer:
	@go run cmd/producer/main.go localhost:9092 important

.PHONY: run-consumer
## run-consumer: Run consumer
run-consumer:
	@go run cmd/consumer/main.go localhost:9092 consumer important

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