package service

import (
	"context"
	"strings"

	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/catalog/repository"
)

type Service struct {
	repository repository.ProductRepository
}

func New(repository repository.ProductRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Product{}, domain.ErrInvalidProduct
	}

	return s.repository.GetByID(ctx, id)
}

func (s *Service) CreateProduct(ctx context.Context, product domain.Product) (domain.Product, error) {
	if strings.TrimSpace(product.ID) == "" ||
		strings.TrimSpace(product.Name) == "" ||
		strings.TrimSpace(product.Currency) == "" ||
		product.PriceCents < 0 {
		return domain.Product{}, domain.ErrInvalidProduct
	}

	return s.repository.Save(ctx, product)
}
