package repository

import (
	"context"
	"sync"

	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/domain"
)

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (domain.Product, error)
	Save(ctx context.Context, product domain.Product) (domain.Product, error)
}

type MemoryRepository struct {
	mu       sync.RWMutex
	products map[string]domain.Product
}

func NewMemoryRepository(seed []domain.Product) *MemoryRepository {
	products := make(map[string]domain.Product, len(seed))
	for _, product := range seed {
		products[product.ID] = product
	}

	return &MemoryRepository{
		products: products,
	}
}

func (r *MemoryRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	select {
	case <-ctx.Done():
		return domain.Product{}, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	product, ok := r.products[id]
	if !ok {
		return domain.Product{}, domain.ErrProductNotFound
	}

	return product, nil
}

func (r *MemoryRepository) Save(ctx context.Context, product domain.Product) (domain.Product, error) {
	select {
	case <-ctx.Done():
		return domain.Product{}, ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[product.ID] = product
	return product, nil
}
