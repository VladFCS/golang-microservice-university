package service

import (
	"context"
	"strings"

	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/repository"
)

type Service struct {
	repository repository.StockRepository
}

func New(repository repository.StockRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetStock(ctx context.Context, productID string) (domain.Stock, error) {
	if strings.TrimSpace(productID) == "" {
		return domain.Stock{}, domain.ErrInvalidReservation
	}

	return s.repository.GetByProductID(ctx, productID)
}

func (s *Service) ReserveStock(ctx context.Context, productID string, quantity int64) (domain.Stock, error) {
	if strings.TrimSpace(productID) == "" || quantity <= 0 {
		return domain.Stock{}, domain.ErrInvalidReservation
	}

	return s.repository.Reserve(ctx, productID, quantity)
}
