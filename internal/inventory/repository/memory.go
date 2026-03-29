package repository

import (
	"context"
	"sync"

	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/domain"
)

type StockRepository interface {
	GetByProductID(ctx context.Context, productID string) (domain.Stock, error)
	Reserve(ctx context.Context, productID string, quantity int64) (domain.Stock, error)
}

type MemoryRepository struct {
	mu     sync.RWMutex
	stocks map[string]domain.Stock
}

func NewMemoryRepository(seed []domain.Stock) *MemoryRepository {
	stocks := make(map[string]domain.Stock, len(seed))
	for _, stock := range seed {
		stocks[stock.ProductID] = stock
	}

	return &MemoryRepository{stocks: stocks}
}

func (r *MemoryRepository) GetByProductID(ctx context.Context, productID string) (domain.Stock, error) {
	select {
	case <-ctx.Done():
		return domain.Stock{}, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	stock, ok := r.stocks[productID]
	if !ok {
		return domain.Stock{}, domain.ErrStockNotFound
	}

	return stock, nil
}

func (r *MemoryRepository) Reserve(ctx context.Context, productID string, quantity int64) (domain.Stock, error) {
	select {
	case <-ctx.Done():
		return domain.Stock{}, ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	stock, ok := r.stocks[productID]
	if !ok {
		return domain.Stock{}, domain.ErrStockNotFound
	}
	if stock.Available < quantity {
		return domain.Stock{}, domain.ErrInsufficientStock
	}

	stock.Available -= quantity
	stock.Reserved += quantity
	r.stocks[productID] = stock

	return stock, nil
}
