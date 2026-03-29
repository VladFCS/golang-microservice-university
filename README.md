# Go Microservices with gRPC and Docker

This repository contains a small production-style microservices system built with Go:

- `gateway-service`: HTTP API that aggregates gRPC responses.
- `catalog-service`: product catalog with in-memory storage.
- `inventory-service`: stock service with in-memory storage.

## Architecture

```text
HTTP Client
    |
    v
gateway-service (HTTP :8080)
    |------------------------------|
    v                              v
catalog-service (gRPC :50051)   inventory-service (gRPC :50052)
```

Each service follows a simple clean architecture flow:

```text
handler -> service -> repository
```

## Project layout

```text
.
├── api/proto/               # Proto definitions
├── cmd/                     # Service entrypoints
├── gen/                     # Generated protobuf Go code
├── internal/
│   ├── catalog/             # catalog-service layers
│   ├── gateway/             # gateway-service layers
│   ├── inventory/           # inventory-service layers
│   └── platform/            # shared logging and middleware
└── deployments/docker/      # Multi-stage Dockerfiles
```

## gRPC contracts

- Catalog
  - `GetProduct`
  - `CreateProduct`
- Inventory
  - `GetStock`
  - `ReserveStock`

Proto files live in:

- `api/proto/catalog/v1/catalog.proto`
- `api/proto/inventory/v1/inventory.proto`

## Generate protobuf code

```bash
make proto
```

## Run locally

In separate terminals:

```bash
make run-catalog
make run-inventory
make run-gateway
```

Fetch aggregated data:

```bash
curl http://localhost:8080/products/p-100
```

Example response:

```json
{
  "id": "p-100",
  "name": "Mechanical Keyboard",
  "description": "Hot-swappable mechanical keyboard",
  "price_cents": 12999,
  "currency": "USD",
  "inventory": {
    "product_id": "p-100",
    "available": 42,
    "reserved": 3
  }
}
```

## gRPC server and client usage

Server registration happens in:

- `cmd/catalog-service/main.go`
- `cmd/inventory-service/main.go`

Client usage happens in the gateway:

- `cmd/gateway-service/main.go`
- `internal/gateway/client/catalog/client.go`
- `internal/gateway/client/inventory/client.go`

You can also call the gRPC servers directly with `grpcurl`:

```bash
grpcurl -plaintext localhost:50051 catalog.v1.CatalogService/GetProduct \
  -d '{"id":"p-100"}'

grpcurl -plaintext localhost:50052 inventory.v1.InventoryService/GetStock \
  -d '{"product_id":"p-100"}'
```

Create a new product:

```bash
grpcurl -plaintext localhost:50051 catalog.v1.CatalogService/CreateProduct \
  -d '{"product":{"id":"p-300","name":"USB-C Dock","description":"11-in-1 dock","price_cents":8999,"currency":"USD"}}'
```

Reserve stock:

```bash
grpcurl -plaintext localhost:50052 inventory.v1.InventoryService/ReserveStock \
  -d '{"product_id":"p-100","quantity":2}'
```

## Build and test

```bash
make tidy
make build
make test
```

## Docker

Build images:

```bash
make docker-build
```

Dockerfiles:

- `deployments/docker/gateway-service.Dockerfile`
- `deployments/docker/catalog-service.Dockerfile`
- `deployments/docker/inventory-service.Dockerfile`
