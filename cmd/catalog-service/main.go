package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	catalogv1 "github.com/vlad/microservices-grpc-kubernetes/gen/catalog/v1"
	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/handler"
	catalogpostgres "github.com/vlad/microservices-grpc-kubernetes/internal/catalog/repository/postgres"
	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/service"
	grpcplatform "github.com/vlad/microservices-grpc-kubernetes/internal/platform/grpcutil"
	"github.com/vlad/microservices-grpc-kubernetes/internal/platform/health"
	"github.com/vlad/microservices-grpc-kubernetes/internal/platform/logger"
	platformpostgres "github.com/vlad/microservices-grpc-kubernetes/internal/platform/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log := logger.New("catalog-service")
	grpcPort := getenv("GRPC_PORT", "50051")
	healthPort := getenv("HEALTH_PORT", "8081")
	databaseURL := getenv("DATABASE_URL", "postgresql://app:app@localhost:5432/microservices?sslmode=disable")

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer startupCancel()

	dbPool, err := platformpostgres.Connect(startupCtx, databaseURL)
	if err != nil {
		log.Error("failed to connect to postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := platformpostgres.RunMigrations(startupCtx, dbPool); err != nil {
		log.Error("failed to run postgres migrations", slog.Any("error", err))
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Error("failed to listen", slog.Any("error", err))
		os.Exit(1)
	}

	repository := catalogpostgres.New(dbPool)
	if err := repository.Seed(startupCtx, []domain.Product{
		{
			ID:          "p-100",
			Name:        "Mechanical Keyboard",
			Description: "Hot-swappable mechanical keyboard",
			PriceCents:  12999,
			Currency:    "USD",
		},
		{
			ID:          "p-200",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			PriceCents:  5999,
			Currency:    "USD",
		},
	}); err != nil {
		log.Error("failed to seed catalog data", slog.Any("error", err))
		os.Exit(1)
	}
	service := service.New(repository)
	grpcHandler := handler.NewGRPCHandler(service, log)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcplatform.UnaryLoggingInterceptor(log)),
	)
	catalogv1.RegisterCatalogServiceServer(server, grpcHandler)
	reflection.Register(server)

	healthServer := &http.Server{
		Addr:              ":" + healthPort,
		Handler:           health.Handler("catalog-service"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("catalog-service started",
			slog.String("grpc_port", grpcPort),
			slog.String("health_port", healthPort),
		)
		if serveErr := server.Serve(lis); serveErr != nil && ctx.Err() == nil {
			log.Error("grpc server stopped with error", slog.Any("error", serveErr))
			stop()
		}
	}()

	go func() {
		if serveErr := healthServer.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Error("health server stopped with error", slog.Any("error", serveErr))
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down catalog-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcStopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(grpcStopped)
	}()

	select {
	case <-grpcStopped:
	case <-shutdownCtx.Done():
		server.Stop()
	}

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown health server gracefully", slog.Any("error", err))
		_ = healthServer.Close()
	}
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}
