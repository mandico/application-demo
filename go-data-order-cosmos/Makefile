NAME=go-data-order-cosmos
VERSION=0.0.1
AUTHOR=mandico

.PHONY: run
## run: Run go-data-order-cosmos
run:
	@go run main.go localhost:9092 consumer_go_data_order order-input

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