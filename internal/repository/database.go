package repository

import (
	"context"
	"fmt"

	"chatmix-backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Client           *mongo.Client
	DB               *mongo.Database
	UserRepo         UserRepository
	RefreshTokenRepo RefreshTokenRepository
	SessionRepo      SessionRepository
	CaptchaRepo      CaptchaRepository
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	clientOptions := options.Client().ApplyURI(cfg.Database.URI)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Database.Name)

	userRepo := NewUserRepository(db, cfg.Database.Collections.Users)
	refreshTokenRepo := NewRefreshTokenRepository(db, cfg.Database.Collections.RefreshTokens)
	sessionRepo := NewSessionRepository(db, cfg.Database.Collections.Sessions)
	captchaRepo := NewCaptchaRepository(db, cfg.Database.Collections.Captchas)

	database := &Database{
		Client:           client,
		DB:               db,
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		SessionRepo:      sessionRepo,
		CaptchaRepo:      captchaRepo,
	}

	// Create indexes
	if err := database.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return database, nil
}

func (d *Database) Close(ctx context.Context) error {
	if d.Client != nil {
		return d.Client.Disconnect(ctx)
	}
	return nil
}

func (d *Database) createIndexes(ctx context.Context) error {
	if userRepo, ok := d.UserRepo.(*userRepository); ok {
		if err := userRepo.CreateIndexes(ctx); err != nil {
			return fmt.Errorf("failed to create user indexes: %w", err)
		}
	}

	if rtRepo, ok := d.RefreshTokenRepo.(*refreshTokenRepository); ok {
		if err := rtRepo.CreateIndexes(ctx); err != nil {
			return fmt.Errorf("failed to create refresh token indexes: %w", err)
		}
	}

	if sessionRepo, ok := d.SessionRepo.(*sessionRepository); ok {
		if err := sessionRepo.CreateIndexes(ctx); err != nil {
			return fmt.Errorf("failed to create session indexes: %w", err)
		}
	}

	if captchaRepo, ok := d.CaptchaRepo.(*captchaRepository); ok {
		if err := captchaRepo.CreateIndexes(ctx); err != nil {
			return fmt.Errorf("failed to create captcha indexes: %w", err)
		}
	}

	return nil
}
