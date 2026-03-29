package service

import (
	"context"
	"testing"

	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/repository"
)

func TestReserveStock(t *testing.T) {
	t.Parallel()

	svc := New(repository.NewMemoryRepository([]domain.Stock{
		{ProductID: "p-100", Available: 10, Reserved: 0},
	}))

	stock, err := svc.ReserveStock(context.Background(), "p-100", 3)
	if err != nil {
		t.Fatalf("ReserveStock() error = %v", err)
	}

	if stock.Available != 7 || stock.Reserved != 3 {
		t.Fatalf("ReserveStock() = %+v, want available=7 reserved=3", stock)
	}
}

func TestReserveStockInsufficientInventory(t *testing.T) {
	t.Parallel()

	svc := New(repository.NewMemoryRepository([]domain.Stock{
		{ProductID: "p-100", Available: 2, Reserved: 0},
	}))

	_, err := svc.ReserveStock(context.Background(), "p-100", 3)
	if err != domain.ErrInsufficientStock {
		t.Fatalf("ReserveStock() error = %v, want %v", err, domain.ErrInsufficientStock)
	}
}
