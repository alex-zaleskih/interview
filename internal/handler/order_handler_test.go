package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/alex-zaleskih/interview/internal/auth"
	"github.com/alex-zaleskih/interview/internal/model"
)

const testUserID = "652f1a000000000000000001"

func init() {
	gin.SetMode(gin.TestMode)
}

type fakeOrderService struct {
	listOrders func(ctx context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error)
}

func (f *fakeOrderService) ListOrders(ctx context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error) {
	return f.listOrders(ctx, userID, page)
}

type fakeUserService struct{}

func (fakeUserService) Login(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("not implemented")
}

func (fakeUserService) GetUser(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("not implemented")
}

type testServer struct {
	router   *gin.Engine
	sessions *auth.Sessions
}

func newTestServer(t *testing.T, orders OrderService) *testServer {
	t.Helper()
	sessions := auth.NewSessions([]byte("test-secret"), time.Hour)
	router := NewRouter(Deps{
		Orders:   orders,
		Users:    fakeUserService{},
		Sessions: sessions,
		WebDir:   t.TempDir(),
	})
	return &testServer{router: router, sessions: sessions}
}

// get performs a GET request on behalf of testUserID.
func (s *testServer) get(target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: s.sessions.Issue(testUserID)})
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func TestOrderHandler_List_ReturnsOrdersOfCurrentUser(t *testing.T) {
	createdAt := time.Date(2026, 9, 28, 14, 30, 0, 0, time.UTC)

	var gotUserID string
	svc := &fakeOrderService{
		listOrders: func(_ context.Context, userID string, _ model.PageRequest) (model.Page[model.OrderSummary], error) {
			gotUserID = userID
			return model.Page[model.OrderSummary]{
				Items: []model.OrderSummary{{
					ID:         "66f000000000000000000abc",
					Number:     "ORD-2026-00042",
					Status:     model.OrderStatusShipped,
					CreatedAt:  createdAt,
					ItemsCount: 3,
					Total:      model.Money{Amount: 1259000, Currency: "RUB"},
				}},
				Total: 21,
				Page:  1,
				Limit: 10,
			}, nil
		},
	}

	rec := newTestServer(t, svc).get("/api/orders")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}
	if gotUserID != testUserID {
		t.Errorf("service called with userID = %q, want %q", gotUserID, testUserID)
	}

	var body orderListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := orderListResponse{
		Items: []orderSummaryResponse{{
			ID:         "66f000000000000000000abc",
			Number:     "ORD-2026-00042",
			Status:     "shipped",
			CreatedAt:  createdAt,
			ItemsCount: 3,
			Total:      moneyResponse{Amount: 1259000, Currency: "RUB"},
		}},
		Total: 21,
		Page:  1,
		Limit: 10,
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("response = %+v, want %+v", body, want)
	}
}

func TestOrderHandler_List_EmptyListIsNotNull(t *testing.T) {
	svc := &fakeOrderService{
		listOrders: func(context.Context, string, model.PageRequest) (model.Page[model.OrderSummary], error) {
			return model.Page[model.OrderSummary]{Page: 1, Limit: 10}, nil
		},
	}

	rec := newTestServer(t, svc).get("/api/orders")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if items, ok := body["items"].([]any); !ok || len(items) != 0 {
		t.Errorf("items = %v, want empty JSON array", body["items"])
	}
}

func TestOrderHandler_List_PassesPagination(t *testing.T) {
	var gotPage model.PageRequest
	svc := &fakeOrderService{
		listOrders: func(_ context.Context, _ string, page model.PageRequest) (model.Page[model.OrderSummary], error) {
			gotPage = page
			return model.Page[model.OrderSummary]{}, nil
		},
	}

	rec := newTestServer(t, svc).get("/api/orders?page=3&limit=5")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if want := (model.PageRequest{Page: 3, Limit: 5}); gotPage != want {
		t.Errorf("page = %+v, want %+v", gotPage, want)
	}
}

func TestOrderHandler_List_InvalidPagination(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "page is not a number", query: "?page=abc"},
		{name: "page is zero", query: "?page=0"},
		{name: "limit is negative", query: "?limit=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeOrderService{
				listOrders: func(context.Context, string, model.PageRequest) (model.Page[model.OrderSummary], error) {
					t.Fatal("service must not be called")
					return model.Page[model.OrderSummary]{}, nil
				},
			}

			rec := newTestServer(t, svc).get("/api/orders" + tt.query)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestOrderHandler_List_ServiceError(t *testing.T) {
	svc := &fakeOrderService{
		listOrders: func(context.Context, string, model.PageRequest) (model.Page[model.OrderSummary], error) {
			return model.Page[model.OrderSummary]{}, errors.New("mongo: connection refused")
		},
	}

	rec := newTestServer(t, svc).get("/api/orders")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != "internal server error" {
		t.Errorf("error = %q, internal details must not leak to the client", body.Error)
	}
}

func TestOrderHandler_List_RequiresAuth(t *testing.T) {
	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "no session cookie", cookie: nil},
		{name: "forged session cookie", cookie: &http.Cookie{Name: sessionCookieName, Value: testUserID + ".9999999999.forged"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeOrderService{
				listOrders: func(context.Context, string, model.PageRequest) (model.Page[model.OrderSummary], error) {
					t.Fatal("service must not be called")
					return model.Page[model.OrderSummary]{}, nil
				},
			}
			srv := newTestServer(t, svc)

			req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()
			srv.router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}
