package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/alex-zaleskih/interview/internal/model"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, errorResponse{Error: message})
}

// respondInternalError logs the real error and hides its details from the client.
func respondInternalError(c *gin.Context, err error) {
	slog.ErrorContext(c.Request.Context(), "request failed",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"error", err,
	)
	respondError(c, http.StatusInternalServerError, "internal server error")
}

type moneyResponse struct {
	Amount   int64  `json:"amount"` // minor units, e.g. kopecks
	Currency string `json:"currency"`
}

func newMoneyResponse(m model.Money) moneyResponse {
	return moneyResponse{Amount: m.Amount, Currency: m.Currency}
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(u model.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
	}
}

type orderSummaryResponse struct {
	ID         string        `json:"id"`
	Number     string        `json:"number"`
	Status     string        `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	ItemsCount int           `json:"items_count"`
	Total      moneyResponse `json:"total"`
}

func newOrderSummaryResponse(o model.OrderSummary) orderSummaryResponse {
	return orderSummaryResponse{
		ID:         o.ID,
		Number:     o.Number,
		Status:     string(o.Status),
		CreatedAt:  o.CreatedAt,
		ItemsCount: o.ItemsCount,
		Total:      newMoneyResponse(o.Total),
	}
}

type orderListResponse struct {
	Items []orderSummaryResponse `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}
