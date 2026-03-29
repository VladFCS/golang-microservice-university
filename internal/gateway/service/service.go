package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/domain"
	"golang.org/x/sync/errgroup"
)

type CatalogRepository interface {
	GetProduct(ctx context.Context, id string) (domain.Product, error)
}

type InventoryRepository interface {
	GetStock(ctx context.Context, productID string) (domain.Stock, error)
}

type Service struct {
	catalogRepo    CatalogRepository
	inventoryRepo  InventoryRepository
	logger         *slog.Logger
	requestTimeout time.Duration
}

func New(
	catalogRepo CatalogRepository,
	inventoryRepo InventoryRepository,
	logger *slog.Logger,
	requestTimeout time.Duration,
) *Service {
	return &Service{
		catalogRepo:    catalogRepo,
		inventoryRepo:  inventoryRepo,
		logger:         logger,
		requestTimeout: requestTimeout,
	}
}

func (s *Service) GetProductDetails(ctx context.Context, id string) (domain.ProductDetails, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.ProductDetails{}, domain.ErrProductIDRequired
	}

	downstreamCtx, cancel := context.WithTimeout(ctx, s.requestTimeout)
	defer cancel()

	s.logger.InfoContext(ctx, "gateway started downstream aggregation",
		slog.String("product_id", id),
		slog.Duration("timeout", s.requestTimeout),
	)

	var (
		product domain.Product
		stock   domain.Stock
	)

	group, groupCtx := errgroup.WithContext(downstreamCtx)
	group.Go(func() error {
		start := time.Now()
		s.logger.InfoContext(groupCtx, "gateway forwarding request to catalog-service",
			slog.String("product_id", id),
		)

		var err error
		product, err = s.catalogRepo.GetProduct(groupCtx, id)
		if err != nil {
			s.logger.ErrorContext(groupCtx, "gateway received error from catalog-service",
				slog.String("product_id", id),
				slog.Duration("duration", time.Since(start)),
				slog.Any("error", err),
			)
			return err
		}

		s.logger.InfoContext(groupCtx, "gateway received response from catalog-service",
			slog.String("product_id", id),
			slog.Duration("duration", time.Since(start)),
			slog.String("name", product.Name),
			slog.Int64("price_cents", product.PriceCents),
		)
		return err
	})
	group.Go(func() error {
		start := time.Now()
		s.logger.InfoContext(groupCtx, "gateway forwarding request to inventory-service",
			slog.String("product_id", id),
		)

		var err error
		stock, err = s.inventoryRepo.GetStock(groupCtx, id)
		if err != nil {
			s.logger.ErrorContext(groupCtx, "gateway received error from inventory-service",
				slog.String("product_id", id),
				slog.Duration("duration", time.Since(start)),
				slog.Any("error", err),
			)
			return err
		}

		s.logger.InfoContext(groupCtx, "gateway received response from inventory-service",
			slog.String("product_id", id),
			slog.Duration("duration", time.Since(start)),
			slog.Int64("available", stock.Available),
			slog.Int64("reserved", stock.Reserved),
		)
		return err
	})

	if err := group.Wait(); err != nil {
		s.logger.ErrorContext(ctx, "gateway aggregation failed",
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		return domain.ProductDetails{}, err
	}

	result := domain.ProductDetails{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Currency:    product.Currency,
		Inventory:   stock,
	}

	s.logger.InfoContext(ctx, "gateway aggregation completed",
		slog.String("product_id", id),
		slog.String("name", result.Name),
		slog.Int64("available", result.Inventory.Available),
	)

	return result, nil
}
