package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/domain"
	db "github.com/vlad/microservices-grpc-kubernetes/internal/catalog/repository/postgres/sqlc"
)

type Repository struct {
	queries *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	row, err := r.queries.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}

		return domain.Product{}, fmt.Errorf("get product by id: %w", err)
	}

	return toDomainProduct(row), nil
}

func (r *Repository) Save(ctx context.Context, product domain.Product) (domain.Product, error) {
	row, err := r.queries.UpsertProduct(ctx, db.UpsertProductParams{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Currency:    product.Currency,
	})
	if err != nil {
		return domain.Product{}, fmt.Errorf("save product: %w", err)
	}

	return toDomainProduct(row), nil
}

func (r *Repository) Seed(ctx context.Context, products []domain.Product) error {
	for _, product := range products {
		if _, err := r.Save(ctx, product); err != nil {
			return fmt.Errorf("seed product %s: %w", product.ID, err)
		}
	}

	return nil
}

func toDomainProduct(row db.Product) domain.Product {
	return domain.Product{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		PriceCents:  row.PriceCents,
		Currency:    row.Currency,
	}
}
