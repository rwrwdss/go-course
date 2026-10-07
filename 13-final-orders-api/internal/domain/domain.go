// Package domain — бизнес-сущности и ошибки. Ни от чего не зависит.
package domain

import "errors"

// Статусы заказа.
type Status string

const (
	StatusNew       Status = "new"
	StatusPaid      Status = "paid"
	StatusShipped   Status = "shipped"
	StatusCancelled Status = "cancelled"
)

// Product — товар каталога. Цена в копейках.
type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

// Item — позиция в заказе.
type Item struct {
	ProductID int   `json:"product_id"`
	Quantity  int   `json:"quantity"`
	Price     int64 `json:"price"` // цена на момент заказа (фиксируем, даже если в каталоге поменяется)
}

// Order — заказ.
type Order struct {
	ID       int    `json:"id"`
	Customer string `json:"customer"`
	Items    []Item `json:"items"`
	Total    int64  `json:"total"`
	Status   Status `json:"status"`
}

// Доменные ошибки. HTTP-слой маппит их в статус-коды.
var (
	ErrNotFound          = errors.New("not found")
	ErrValidation        = errors.New("validation error")
	ErrInvalidTransition = errors.New("invalid status transition")
)
