package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/example/shop-cabinet/internal/model"
)

type UserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{coll: db.Collection(usersCollection)}
}

type userDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	Email     string        `bson:"email"`
	FullName  string        `bson:"full_name"`
	Phone     string        `bson:"phone"`
	CreatedAt time.Time     `bson:"created_at"`
}

func (d userDocument) toModel() model.User {
	return model.User{
		ID:        d.ID.Hex(),
		Email:     d.Email,
		FullName:  d.FullName,
		Phone:     d.Phone,
		CreatedAt: d.CreatedAt,
	}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (model.User, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return model.User{}, model.ErrNotFound
	}
	return r.findOne(ctx, bson.D{{Key: "_id", Value: oid}})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	return r.findOne(ctx, bson.D{{Key: "email", Value: email}})
}

func (r *UserRepository) findOne(ctx context.Context, filter bson.D) (model.User, error) {
	var doc userDocument
	err := r.coll.FindOne(ctx, filter).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.User{}, model.ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("find user: %w", err)
	}
	return doc.toModel(), nil
}
