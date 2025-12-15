package repository

import (
	"context"
	"register/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRepo struct {
	coll *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *mongoRepo {
	return &mongoRepo{coll: db.Collection("users")}
}

func (r *mongoRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = primitive.NewObjectID().Hex()
	_, err := r.coll.InsertOne(ctx, user)
	return err
}

func (r *mongoRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (r *mongoRepo) List(ctx context.Context) ([]*model.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *mongoRepo) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

func (r *mongoRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mongoRepo) Update(ctx context.Context, id, name, email string) (*model.User, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.User
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"name": name, "email": email},
	}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *mongoRepo) Delete(ctx context.Context, id string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
