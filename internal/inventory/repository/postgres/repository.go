package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/domain"
	db "github.com/vlad/microservices-grpc-kubernetes/internal/inventory/repository/postgres/sqlc"
)

type Repository struct {
	queries *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func (r *Repository) GetByProductID(ctx context.Context, productID string) (domain.Stock, error) {
	row, err := r.queries.GetStockByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Stock{}, domain.ErrStockNotFound
		}

		return domain.Stock{}, fmt.Errorf("get stock by product id: %w", err)
	}

	return toDomainStock(row), nil
}

func (r *Repository) Reserve(ctx context.Context, productID string, quantity int64) (domain.Stock, error) {
	row, err := r.queries.ReserveStock(ctx, db.ReserveStockParams{
		ProductID: productID,
		Available: quantity,
	})
	if err == nil {
		return toDomainStock(row), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Stock{}, fmt.Errorf("reserve stock: %w", err)
	}

	current, getErr := r.GetByProductID(ctx, productID)
	if getErr != nil {
		return domain.Stock{}, getErr
	}
	if current.Available < quantity {
		return domain.Stock{}, domain.ErrInsufficientStock
	}

	return domain.Stock{}, fmt.Errorf("reserve stock: update returned no rows")
}

func (r *Repository) Seed(ctx context.Context, stocks []domain.Stock) error {
	for _, stock := range stocks {
		if _, err := r.queries.UpsertStock(ctx, db.UpsertStockParams{
			ProductID: stock.ProductID,
			Available: stock.Available,
			Reserved:  stock.Reserved,
		}); err != nil {
			return fmt.Errorf("seed stock %s: %w", stock.ProductID, err)
		}
	}

	return nil
}

func toDomainStock(row db.InventoryStock) domain.Stock {
	return domain.Stock{
		ProductID: row.ProductID,
		Available: row.Available,
		Reserved:  row.Reserved,
	}
}
