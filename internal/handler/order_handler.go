package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/shop-cabinet/internal/model"
)

type OrderService interface {
	ListOrders(ctx context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error)
}

type OrderHandler struct {
	orders OrderService
}

func NewOrderHandler(orders OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

// List handles GET /api/orders?page=1&limit=10 and returns orders of the current user.
func (h *OrderHandler) List(c *gin.Context) {
	page, err := positiveIntQuery(c, "page")
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	limit, err := positiveIntQuery(c, "limit")
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.orders.ListOrders(c.Request.Context(), currentUserID(c), model.PageRequest{Page: page, Limit: limit})
	if err != nil {
		respondInternalError(c, err)
		return
	}

	resp := orderListResponse{
		Items: make([]orderSummaryResponse, 0, len(result.Items)),
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	}
	for _, o := range result.Items {
		resp.Items = append(resp.Items, newOrderSummaryResponse(o))
	}
	c.JSON(http.StatusOK, resp)
}

// positiveIntQuery parses an optional positive integer query parameter. A missing parameter yields 0.
func positiveIntQuery(c *gin.Context, key string) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return 0, fmt.Errorf("query parameter %q must be a positive integer", key)
	}
	return v, nil
}
