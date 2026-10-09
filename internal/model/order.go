package model

import "time"

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Money is an amount in minor currency units (e.g. kopecks for RUB).
type Money struct {
	Amount   int64
	Currency string
}

// OrderSummary is a short representation of an order used in the order list.
type OrderSummary struct {
	ID         string
	Number     string
	Status     OrderStatus
	CreatedAt  time.Time
	ItemsCount int
	Total      Money
}
