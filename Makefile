.PHONY: help build test lint clean install run docker

help:
build:
build-cli:
test:
test-unit:
test-integration:
test-coverage:
lint:
fmt:
clean:
install:
docker:
docker-compose:
deps:
tidy:
proto:
	protoc --go_out=. --go-grpc_out=. api/proto/*.proto
release:
dev: