#!make

GOPATH := $(shell go env GOPATH)
CID := $(shell docker ps -q -f "name=rabbitmq")
export RABBIT_MQ_PORT := 5672
export QUEUE_NAME := bloxroute
export USERNAME := guest
export PASSWORD := guest
export CLIENT_INPUT_FILENAME := input.json
export SERVER_OUTPUT_FILENAME := output.json

.PHONY: client
client:
	go run cmd/client/main.go

.PHONY: server
server:
	go run cmd/server/main.go

.PHONY: test
test:
	go clean -testcache
	go test -v ./... -race


.PHONY: build-env
build-env:
ifeq ($(CID),)
	docker run -d --rm --name rabbitmq -p $(RABBIT_MQ_PORT):$(RABBIT_MQ_PORT) rabbitmq
endif



