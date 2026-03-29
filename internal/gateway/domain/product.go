package domain

import "errors"

var (
	ErrProductIDRequired = errors.New("product id is required")
	ErrProductNotFound   = errors.New("product not found")
	ErrStockNotFound     = errors.New("stock not found")
)

type Product struct {
	ID          string
	Name        string
	Description string
	PriceCents  int64
	Currency    string
}

type Stock struct {
	ProductID string `json:"product_id"`
	Available int64  `json:"available"`
	Reserved  int64  `json:"reserved"`
}

type ProductDetails struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Currency    string `json:"currency"`
	Inventory   Stock  `json:"inventory"`
}
