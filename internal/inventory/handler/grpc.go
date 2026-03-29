package handler

import (
	"context"
	"errors"
	"log/slog"

	inventoryv1 "github.com/vlad/microservices-grpc-kubernetes/gen/inventory/v1"
	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/inventory/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	inventoryv1.UnimplementedInventoryServiceServer
	service *service.Service
	logger  *slog.Logger
}

func NewGRPCHandler(service *service.Service, logger *slog.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GRPCHandler) GetStock(
	ctx context.Context,
	req *inventoryv1.GetStockRequest,
) (*inventoryv1.GetStockResponse, error) {
	stock, err := h.service.GetStock(ctx, req.GetProductId())
	if err != nil {
		return nil, mapInventoryError(err)
	}

	return &inventoryv1.GetStockResponse{
		Stock: toProtoStock(stock),
	}, nil
}

func (h *GRPCHandler) ReserveStock(
	ctx context.Context,
	req *inventoryv1.ReserveStockRequest,
) (*inventoryv1.ReserveStockResponse, error) {
	stock, err := h.service.ReserveStock(ctx, req.GetProductId(), req.GetQuantity())
	if err != nil {
		return nil, mapInventoryError(err)
	}

	h.logger.InfoContext(ctx, "stock reserved",
		slog.String("product_id", stock.ProductID),
		slog.Int64("reserved", req.GetQuantity()),
	)

	return &inventoryv1.ReserveStockResponse{
		Stock: toProtoStock(stock),
	}, nil
}

func toProtoStock(stock domain.Stock) *inventoryv1.Stock {
	return &inventoryv1.Stock{
		ProductId: stock.ProductID,
		Available: stock.Available,
		Reserved:  stock.Reserved,
	}
}

func mapInventoryError(err error) error {
	switch {
	case errors.Is(err, domain.ErrStockNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidReservation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
