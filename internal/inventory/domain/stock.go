package domain

import "errors"

var (
	ErrStockNotFound      = errors.New("stock not found")
	ErrInvalidReservation = errors.New("invalid reservation")
	ErrInsufficientStock  = errors.New("insufficient stock")
)

type Stock struct {
	ProductID string
	Available int64
	Reserved  int64
}
