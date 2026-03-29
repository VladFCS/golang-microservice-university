package inventory

import (
	"context"
	"errors"

	inventoryv1 "github.com/vlad/microservices-grpc-kubernetes/gen/inventory/v1"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client inventoryv1.InventoryServiceClient
}

func New(client inventoryv1.InventoryServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) GetStock(ctx context.Context, productID string) (domain.Stock, error) {
	resp, err := c.client.GetStock(ctx, &inventoryv1.GetStockRequest{ProductId: productID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.Stock{}, domain.ErrStockNotFound
		}

		return domain.Stock{}, err
	}
	if resp.GetStock() == nil {
		return domain.Stock{}, errors.New("inventory response missing stock")
	}

	return domain.Stock{
		ProductID: resp.GetStock().GetProductId(),
		Available: resp.GetStock().GetAvailable(),
		Reserved:  resp.GetStock().GetReserved(),
	}, nil
}
