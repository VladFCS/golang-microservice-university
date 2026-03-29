FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/inventory-service ./cmd/inventory-service

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder --chown=10001:10001 /out/inventory-service /app/inventory-service

USER 10001:10001

EXPOSE 50052 8082

ENTRYPOINT ["/app/inventory-service"]
