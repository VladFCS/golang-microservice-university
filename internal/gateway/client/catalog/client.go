package catalog

import (
	"context"
	"errors"

	catalogv1 "github.com/vlad/microservices-grpc-kubernetes/gen/catalog/v1"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client catalogv1.CatalogServiceClient
}

func New(client catalogv1.CatalogServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	resp, err := c.client.GetProduct(ctx, &catalogv1.GetProductRequest{Id: id})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.Product{}, domain.ErrProductNotFound
		}

		return domain.Product{}, err
	}
	if resp.GetProduct() == nil {
		return domain.Product{}, errors.New("catalog response missing product")
	}

	return domain.Product{
		ID:          resp.GetProduct().GetId(),
		Name:        resp.GetProduct().GetName(),
		Description: resp.GetProduct().GetDescription(),
		PriceCents:  resp.GetProduct().GetPriceCents(),
		Currency:    resp.GetProduct().GetCurrency(),
	}, nil
}
