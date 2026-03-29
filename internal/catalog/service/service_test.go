package service

import (
	"context"
	"testing"

	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/repository"
)

func TestCreateAndGetProduct(t *testing.T) {
	t.Parallel()

	repo := repository.NewMemoryRepository(nil)
	svc := New(repo)

	product, err := svc.CreateProduct(context.Background(), domain.Product{
		ID:          "p-300",
		Name:        "USB-C Dock",
		Description: "11-in-1 dock",
		PriceCents:  8999,
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	got, err := svc.GetProduct(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}

	if got.Name != product.Name {
		t.Fatalf("GetProduct().Name = %q, want %q", got.Name, product.Name)
	}
}

func TestCreateProductRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	svc := New(repository.NewMemoryRepository(nil))

	_, err := svc.CreateProduct(context.Background(), domain.Product{
		ID:         "",
		Name:       "Broken Product",
		Currency:   "USD",
		PriceCents: 100,
	})
	if err != domain.ErrInvalidProduct {
		t.Fatalf("CreateProduct() error = %v, want %v", err, domain.ErrInvalidProduct)
	}
}
