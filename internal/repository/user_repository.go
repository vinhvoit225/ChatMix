package repository

import (
	"context"
	"errors"
	"time"

	"chatmix-backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLastSeen(ctx context.Context, username string) error
	SetOnlineStatus(ctx context.Context, username string, online bool) error
	GetOnlineUsers(ctx context.Context) ([]*model.User, error)
	GetAllUsers(ctx context.Context) ([]*model.User, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByUsername(ctx context.Context, username string) error
	Exists(ctx context.Context, username string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database, collectionName string) UserRepository {
	return &userRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	if user.JoinedAt.IsZero() {
		user.JoinedAt = time.Now()
	}

	if user.LastSeen.IsZero() {
		user.LastSeen = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) UpdateLastSeen(ctx context.Context, username string) error {
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"last_seen": time.Now()}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) SetOnlineStatus(ctx context.Context, username string, online bool) error {
	filter := bson.M{"username": username}
	updateDoc := bson.M{"is_online": online}

	if online {
		updateDoc["last_seen"] = time.Now()
	}

	update := bson.M{"$set": updateDoc}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) GetOnlineUsers(ctx context.Context) ([]*model.User, error) {
	filter := bson.M{"is_online": true}
	opts := options.Find().SetSort(bson.D{{Key: "username", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "joined_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *userRepository) DeleteByUsername(ctx context.Context, username string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"username": username})
	return err
}

func (r *userRepository) Exists(ctx context.Context, username string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *userRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_online", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "last_seen", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "joined_at", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
