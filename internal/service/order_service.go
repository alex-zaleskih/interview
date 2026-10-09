package service

import (
	"context"
	"fmt"

	"github.com/example/shop-cabinet/internal/model"
)

const (
	DefaultPageLimit = 10
	MaxPageLimit     = 50
)

type OrderRepository interface {
	ListByUser(ctx context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error)
}

type OrderService struct {
	orders OrderRepository
}

func NewOrderService(orders OrderRepository) *OrderService {
	return &OrderService{orders: orders}
}

// ListOrders returns a page of orders that belong to the user.
// Missing pagination values are replaced with defaults, the limit is capped by MaxPageLimit.
func (s *OrderService) ListOrders(ctx context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error) {
	result, err := s.orders.ListByUser(ctx, userID, normalizePage(page))
	if err != nil {
		return model.Page[model.OrderSummary]{}, fmt.Errorf("list orders: %w", err)
	}
	return result, nil
}

func normalizePage(p model.PageRequest) model.PageRequest {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = DefaultPageLimit
	}
	if p.Limit > MaxPageLimit {
		p.Limit = MaxPageLimit
	}
	return p
}
