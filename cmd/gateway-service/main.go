package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	catalogv1 "github.com/vlad/microservices-grpc-kubernetes/gen/catalog/v1"
	inventoryv1 "github.com/vlad/microservices-grpc-kubernetes/gen/inventory/v1"
	catalogclient "github.com/vlad/microservices-grpc-kubernetes/internal/gateway/client/catalog"
	inventoryclient "github.com/vlad/microservices-grpc-kubernetes/internal/gateway/client/inventory"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/handler"
	gatewayservice "github.com/vlad/microservices-grpc-kubernetes/internal/gateway/service"
	"github.com/vlad/microservices-grpc-kubernetes/internal/platform/httputil"
	"github.com/vlad/microservices-grpc-kubernetes/internal/platform/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logger.New("gateway-service")
	httpPort := getenv("HTTP_PORT", "8080")
	catalogAddr := getenv("CATALOG_GRPC_ADDR", "localhost:50051")
	inventoryAddr := getenv("INVENTORY_GRPC_ADDR", "localhost:50052")
	requestTimeout := mustDuration(getenv("DOWNSTREAM_TIMEOUT", "3s"))

	catalogConn, err := dialGRPC(context.Background(), catalogAddr)
	if err != nil {
		log.Error("failed to connect to catalog-service", slog.Any("error", err))
		os.Exit(1)
	}
	defer catalogConn.Close()

	inventoryConn, err := dialGRPC(context.Background(), inventoryAddr)
	if err != nil {
		log.Error("failed to connect to inventory-service", slog.Any("error", err))
		os.Exit(1)
	}
	defer inventoryConn.Close()

	service := gatewayservice.New(
		catalogclient.New(catalogv1.NewCatalogServiceClient(catalogConn)),
		inventoryclient.New(inventoryv1.NewInventoryServiceClient(inventoryConn)),
		log,
		requestTimeout,
	)
	httpHandler := handler.NewHTTPHandler(service, log)

	router := chi.NewRouter()
	httpHandler.Register(router)
	server := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           httputil.LoggingMiddleware(log, router),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("gateway-service started",
			slog.String("http_port", httpPort),
			slog.String("catalog_grpc_addr", catalogAddr),
			slog.String("inventory_grpc_addr", inventoryAddr),
		)
		if serveErr := server.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Error("http server stopped with error", slog.Any("error", serveErr))
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down gateway-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown http server gracefully", slog.Any("error", err))
		server.Close()
	}
}

func dialGRPC(ctx context.Context, address string) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return conn, nil
		}

		if !conn.WaitForStateChange(dialCtx, state) {
			_ = conn.Close()
			return nil, dialCtx.Err()
		}
	}
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

func mustDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		panic(err)
	}

	return duration
}
