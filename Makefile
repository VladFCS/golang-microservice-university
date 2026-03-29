MODULE := github.com/vlad/microservices-grpc-kubernetes
PROTO_FILES := api/proto/catalog/v1/catalog.proto api/proto/inventory/v1/inventory.proto
GOBIN := $(shell go env GOPATH)/bin

.PHONY: proto tidy build test run-gateway run-catalog run-inventory docker-build

proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	PATH="$(GOBIN):$(PATH)" protoc \
		--proto_path=api/proto \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

tidy:
	go mod tidy

build:
	go build ./...

test:
	go test ./...

run-gateway:
	go run ./cmd/gateway-service

run-catalog:
	go run ./cmd/catalog-service

run-inventory:
	go run ./cmd/inventory-service

docker-build:
	docker build -f deployments/docker/gateway-service.Dockerfile -t gateway-service:latest .
	docker build -f deployments/docker/catalog-service.Dockerfile -t catalog-service:latest .
	docker build -f deployments/docker/inventory-service.Dockerfile -t inventory-service:latest .
