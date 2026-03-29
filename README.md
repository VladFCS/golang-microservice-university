# Go Microservices with gRPC and Docker

This repository contains a small production-style microservices system built with Go:

- `gateway-service`: HTTP API that aggregates gRPC responses.
- `catalog-service`: product catalog with in-memory storage.
- `inventory-service`: stock service with in-memory storage.

## Repo overview

| Name | Port(s) | Entrypoint | Dockerfile | Path | Dependencies |
| --- | --- | --- | --- | --- | --- |
| `gateway-service` | `8080` HTTP | `cmd/gateway-service/main.go` | `deployments/docker/gateway-service.Dockerfile` | `internal/gateway` | `catalog-service:50051`, `inventory-service:50052` |
| `catalog-service` | `50051` gRPC, `8081` health | `cmd/catalog-service/main.go` | `deployments/docker/catalog-service.Dockerfile` | `internal/catalog` | none |
| `inventory-service` | `50052` gRPC, `8082` health | `cmd/inventory-service/main.go` | `deployments/docker/inventory-service.Dockerfile` | `internal/inventory` | none |

Container baseline for every service:

- multi-stage Docker build
- distroless final image
- non-root runtime with `USER 10001:10001`
- `/healthz` endpoint ready for Kubernetes probes

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

Health endpoints:

- `gateway-service`: `http://localhost:8080/healthz`
- `catalog-service`: `http://localhost:8081/healthz`
- `inventory-service`: `http://localhost:8082/healthz`

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
make docker-build IMAGE_TAG=dev
```

Dockerfiles:

- `deployments/docker/gateway-service.Dockerfile`
- `deployments/docker/catalog-service.Dockerfile`
- `deployments/docker/inventory-service.Dockerfile`

Why this container setup is a good baseline for Kubernetes:

- Multi-stage builds keep the runtime image small and avoid shipping the Go toolchain.
- Distroless runtime images reduce attack surface because they do not include a shell or package manager.
- Containers run as `USER 10001:10001`, which fits common `runAsNonRoot` Kubernetes policies.
- Each service exposes `/healthz`, which can be reused later for readiness and liveness probes.

Published image strategy:

- Registry: GHCR (`ghcr.io`)
- Mutable convenience tags like `latest` are intentionally not used.
- Main branch pushes publish `sha-<commit>` tags.
- Git tags like `v1.0.0` publish matching semver image tags.

GitHub Actions workflow:

- `.github/workflows/publish-images.yml`

Example image names:

- `ghcr.io/<owner>/<repo>-gateway-service:sha-abc1234`
- `ghcr.io/<owner>/<repo>-catalog-service:v1.0.0`
- `ghcr.io/<owner>/<repo>-inventory-service:sha-abc1234`

## Kubernetes

Minimal manifests for a local cluster live in:

- `deploy/k8s/namespace.yaml`
- `deploy/k8s/gateway-deployment.yaml`
- `deploy/k8s/gateway-service.yaml`
- `deploy/k8s/catalog-deployment.yaml`
- `deploy/k8s/catalog-service.yaml`
- `deploy/k8s/inventory-deployment.yaml`
- `deploy/k8s/inventory-service.yaml`
- `deploy/k8s/kustomization.yaml`

Apply them with:

```bash
kubectl apply -k deploy/k8s
```

Kubernetes baseline included in every Deployment:

- consistent `app.kubernetes.io/*` labels and selectors
- readiness and liveness probes
- CPU and memory requests/limits
- `runAsNonRoot: true`
- `allowPrivilegeEscalation: false`
- `readOnlyRootFilesystem: true`
