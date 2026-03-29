FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog-service ./cmd/catalog-service

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder --chown=10001:10001 /out/catalog-service /app/catalog-service

USER 10001:10001

EXPOSE 50051 8081

ENTRYPOINT ["/app/catalog-service"]
