#!make

GOPATH := $(shell go env GOPATH)
CID := $(shell docker ps -q -f "name=rabbitmq")
export RABBIT_MQ_PORT := 5672
export QUEUE_NAME := bloxroute
export USERNAME := guest
export PASSWORD := guest
export CLIENT_INPUT_FILENAME := input.json

.PHONY: client
client:
	go run cmd/client/main.go
	$(GOPATH)/bin/client


.PHONY: server
server:
	go run cmd/server/main.go
	$(GOPATH)/bin/server


.PHONY: build-env
build-env:
ifeq ($(CID),)
	docker run -d --rm --name rabbitmq -p $(RABBIT_MQ_PORT):$(RABBIT_MQ_PORT) rabbitmq
endif



