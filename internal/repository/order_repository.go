package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/alex-zaleskih/interview/internal/model"
)

type OrderRepository struct {
	coll *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{coll: db.Collection(ordersCollection)}
}

// orderSummaryDocument contains only the fields loaded by orderSummaryProjection.
type orderSummaryDocument struct {
	ID         bson.ObjectID `bson:"_id"`
	Number     string        `bson:"number"`
	Status     string        `bson:"status"`
	CreatedAt  time.Time     `bson:"created_at"`
	ItemsCount int           `bson:"items_count"`
	Total      int64         `bson:"total"`
	Currency   string        `bson:"currency"`
}

func (d orderSummaryDocument) toModel() model.OrderSummary {
	return model.OrderSummary{
		ID:         d.ID.Hex(),
		Number:     d.Number,
		Status:     model.OrderStatus(d.Status),
		CreatedAt:  d.CreatedAt,
		ItemsCount: d.ItemsCount,
		Total:      model.Money{Amount: d.Total, Currency: d.Currency},
	}
}

// orderSummaryProjection loads only what the order list needs instead of whole order documents.
var orderSummaryProjection = bson.D{
	{Key: "number", Value: 1},
	{Key: "status", Value: 1},
	{Key: "created_at", Value: 1},
	{Key: "total", Value: 1},
	{Key: "currency", Value: 1},
	{Key: "items_count", Value: bson.D{{Key: "$size", Value: "$items"}}},
}

// ListByUser returns a page of the user's orders, newest first.
func (r *OrderRepository) ListByUser(ctx context.Context, userID string, page model.PageRequest) (model.Page[model.OrderSummary], error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return model.Page[model.OrderSummary]{}, fmt.Errorf("invalid user id %q: %w", userID, err)
	}
	filter := bson.D{{Key: "user_id", Value: uid}}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return model.Page[model.OrderSummary]{}, fmt.Errorf("count orders: %w", err)
	}

	opts := options.Find().
		SetProjection(orderSummaryProjection).
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(page.Offset()).
		SetLimit(int64(page.Limit))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return model.Page[model.OrderSummary]{}, fmt.Errorf("find orders: %w", err)
	}

	var docs []orderSummaryDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return model.Page[model.OrderSummary]{}, fmt.Errorf("decode orders: %w", err)
	}

	items := make([]model.OrderSummary, 0, len(docs))
	for _, doc := range docs {
		items = append(items, doc.toModel())
	}

	return model.Page[model.OrderSummary]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Limit: page.Limit,
	}, nil
}
