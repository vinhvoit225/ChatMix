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

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.RefreshToken, error)
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.RefreshToken, error)
	Update(ctx context.Context, token *model.RefreshToken) error
	Revoke(ctx context.Context, id primitive.ObjectID) error
	RevokeAllByUserID(ctx context.Context, userID primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.Session, error)
	GetByToken(ctx context.Context, token string) (*model.Session, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	DeactivateByToken(ctx context.Context, token string) error
	DeactivateAllByUserID(ctx context.Context, userID primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
}

type CaptchaRepository interface {
	Create(ctx context.Context, captcha *model.CaptchaChallenge) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*model.CaptchaChallenge, error)
	Update(ctx context.Context, captcha *model.CaptchaChallenge) error
	DeleteExpired(ctx context.Context) error
	DeleteByIPAddress(ctx context.Context, ipAddress string) error
}

type refreshTokenRepository struct {
	collection *mongo.Collection
}

type sessionRepository struct {
	collection *mongo.Collection
}

type captchaRepository struct {
	collection *mongo.Collection
}

func NewRefreshTokenRepository(db *mongo.Database, collectionName string) RefreshTokenRepository {
	return &refreshTokenRepository{
		collection: db.Collection(collectionName),
	}
}

func NewSessionRepository(db *mongo.Database, collectionName string) SessionRepository {
	return &sessionRepository{
		collection: db.Collection(collectionName),
	}
}

func NewCaptchaRepository(db *mongo.Database, collectionName string) CaptchaRepository {
	return &captchaRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	if token.ID.IsZero() {
		token.ID = primitive.NewObjectID()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *refreshTokenRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&refreshToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.RefreshToken, error) {
	filter := bson.M{"user_id": userID, "is_revoked": false}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*model.RefreshToken
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *model.RefreshToken) error {
	filter := bson.M{"_id": token.ID}
	update := bson.M{"$set": token}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_revoked": true}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"is_revoked": true}}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"is_revoked": true},
		},
	}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	if session.ID.IsZero() {
		session.ID = primitive.NewObjectID()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.LastUsed.IsZero() {
		session.LastUsed = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *sessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Session, error) {
	var session model.Session
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*model.Session, error) {
	var session model.Session
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.Session, error) {
	filter := bson.M{"user_id": userID, "is_active": true}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*model.Session
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *model.Session) error {
	filter := bson.M{"_id": session.ID}
	update := bson.M{"$set": session}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *sessionRepository) DeactivateByToken(ctx context.Context, token string) error {
	filter := bson.M{"token": token}
	update := bson.M{"$set": bson.M{"is_active": false}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *sessionRepository) DeactivateAllByUserID(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"is_active": false}}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"is_active": false},
		},
	}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *captchaRepository) Create(ctx context.Context, captcha *model.CaptchaChallenge) error {
	if captcha.ID.IsZero() {
		captcha.ID = primitive.NewObjectID()
	}
	if captcha.CreatedAt.IsZero() {
		captcha.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, captcha)
	return err
}

func (r *captchaRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.CaptchaChallenge, error) {
	var captcha model.CaptchaChallenge
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&captcha)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &captcha, nil
}

func (r *captchaRepository) Update(ctx context.Context, captcha *model.CaptchaChallenge) error {
	filter := bson.M{"_id": captcha.ID}
	update := bson.M{"$set": captcha}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *captchaRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"is_used": true},
		},
	}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *captchaRepository) DeleteByIPAddress(ctx context.Context, ipAddress string) error {
	filter := bson.M{"ip_address": ipAddress}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *refreshTokenRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_revoked", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *sessionRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *captchaRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "ip_address", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_used", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
