package service

import (
	"context"
	"errors"
	"testing"

	"github.com/alex-zaleskih/interview/internal/model"
)

type fakeOrderRepository struct {
	gotUserID string
	gotPage   model.PageRequest
	err       error
}

func (f *fakeOrderRepository) ListByUser(_ context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error) {
	f.gotUserID = userID
	f.gotPage = page
	if f.err != nil {
		return model.Page[model.OrderSummary]{}, f.err
	}
	return model.Page[model.OrderSummary]{Page: page.Page, Limit: page.Limit}, nil
}

func TestOrderService_ListOrders_NormalizesPagination(t *testing.T) {
	tests := []struct {
		name string
		in   model.PageRequest
		want model.PageRequest
	}{
		{name: "defaults when not set", in: model.PageRequest{}, want: model.PageRequest{Page: 1, Limit: DefaultPageLimit}},
		{name: "keeps valid values", in: model.PageRequest{Page: 2, Limit: 20}, want: model.PageRequest{Page: 2, Limit: 20}},
		{name: "caps limit", in: model.PageRequest{Page: 1, Limit: 1000}, want: model.PageRequest{Page: 1, Limit: MaxPageLimit}},
		{name: "fixes negative values", in: model.PageRequest{Page: -1, Limit: -1}, want: model.PageRequest{Page: 1, Limit: DefaultPageLimit}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderRepository{}
			svc := NewOrderService(repo)

			if _, err := svc.ListOrders(context.Background(), "user-1", tt.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.gotPage != tt.want {
				t.Errorf("repository called with page %+v, want %+v", repo.gotPage, tt.want)
			}
			if repo.gotUserID != "user-1" {
				t.Errorf("repository called with userID %q, want %q", repo.gotUserID, "user-1")
			}
		})
	}
}

func TestOrderService_ListOrders_WrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("boom")
	svc := NewOrderService(&fakeOrderRepository{err: repoErr})

	_, err := svc.ListOrders(context.Background(), "user-1", model.PageRequest{})

	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want wrapped %v", err, repoErr)
	}
}
