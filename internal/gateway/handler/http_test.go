package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/service"
)

type stubCatalogRepository struct{}

func (stubCatalogRepository) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	return domain.Product{
		ID:          id,
		Name:        "Mechanical Keyboard",
		Description: "Hot-swappable mechanical keyboard",
		PriceCents:  12999,
		Currency:    "USD",
	}, nil
}

type stubInventoryRepository struct{}

func (stubInventoryRepository) GetStock(ctx context.Context, productID string) (domain.Stock, error) {
	return domain.Stock{
		ProductID: productID,
		Available: 42,
		Reserved:  3,
	}, nil
}

func TestGetProductReturnsAggregatedResponse(t *testing.T) {
	t.Parallel()

	svc := service.New(stubCatalogRepository{}, stubInventoryRepository{}, slog.Default(), time.Second)
	handler := NewHTTPHandler(svc, slog.Default())

	router := chi.NewRouter()
	handler.Register(router)

	req := httptest.NewRequest(http.MethodGet, "/products/p-100", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response domain.ProductDetails
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.ID != "p-100" {
		t.Fatalf("response.ID = %q, want %q", response.ID, "p-100")
	}
	if response.Inventory.Available != 42 {
		t.Fatalf("response.Inventory.Available = %d, want %d", response.Inventory.Available, 42)
	}
}

func TestHealthzReturnsOK(t *testing.T) {
	t.Parallel()

	svc := service.New(stubCatalogRepository{}, stubInventoryRepository{}, slog.Default(), time.Second)
	handler := NewHTTPHandler(svc, slog.Default())

	router := chi.NewRouter()
	handler.Register(router)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response["status"] != "ok" {
		t.Fatalf("response[status] = %q, want %q", response["status"], "ok")
	}
}
