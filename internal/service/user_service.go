package service

import (
	"context"
	"fmt"
	"strings"

	"chatmix-backend/internal/config"
	"chatmix-backend/internal/model"
	"chatmix-backend/internal/repository"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	CreateUser(ctx context.Context, username string) (*model.User, error)
	GetUser(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	SetUserOnline(ctx context.Context, username string) error
	SetUserOffline(ctx context.Context, username string) error
	GetOnlineUsers(ctx context.Context) ([]*model.User, error)
	GetAllUsers(ctx context.Context) ([]*model.User, error)
	DeleteUser(ctx context.Context, username string) error
	UserExists(ctx context.Context, username string) (bool, error)
	ValidateUsername(username string) error
	GetUserStats(ctx context.Context) (map[string]interface{}, error)
}

type userService struct {
	userRepo repository.UserRepository
	config   *config.Config
	logger   *logrus.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	config *config.Config,
	logger *logrus.Logger,
) UserService {
	return &userService{
		userRepo: userRepo,
		config:   config,
		logger:   logger,
	}
}

func (s *userService) CreateUser(ctx context.Context, username string) (*model.User, error) {
	if err := s.ValidateUsername(username); err != nil {
		return nil, err
	}

	username = s.sanitizeUsername(username)

	exists, err := s.userRepo.Exists(ctx, username)
	if err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to check if user exists")
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}

	if exists {
		user, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing user: %w", err)
		}

		// Set user online
		if err := s.SetUserOnline(ctx, username); err != nil {
			s.logger.WithError(err).WithField("username", username).Error("Failed to set existing user online")
		}

		return user, nil
	}

	user := model.NewUser(username, username+"@example.com")

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to create user")
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID.Hex(),
		"username": username,
	}).Info("User created successfully")

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, username string) (*model.User, error) {
	if err := s.ValidateUsername(username); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", id.Hex()).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *model.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if err := s.ValidateUsername(user.Username); err != nil {
		return err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  user.ID.Hex(),
			"username": user.Username,
		}).Error("Failed to update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
	}).Info("User updated successfully")

	return nil
}

func (s *userService) SetUserOnline(ctx context.Context, username string) error {
	if err := s.ValidateUsername(username); err != nil {
		return err
	}

	if err := s.userRepo.SetOnlineStatus(ctx, username, true); err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to set user online")
		return fmt.Errorf("failed to set user online: %w", err)
	}

	s.logger.WithField("username", username).Info("User set to online")
	return nil
}

func (s *userService) SetUserOffline(ctx context.Context, username string) error {
	if err := s.ValidateUsername(username); err != nil {
		return err
	}

	if err := s.userRepo.SetOnlineStatus(ctx, username, false); err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to set user offline")
		return fmt.Errorf("failed to set user offline: %w", err)
	}

	s.logger.WithField("username", username).Info("User set to offline")
	return nil
}

func (s *userService) GetOnlineUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.userRepo.GetOnlineUsers(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get online users")
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}

	s.logger.WithField("count", len(users)).Info("Retrieved online users")
	return users, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get all users")
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	s.logger.WithField("count", len(users)).Info("Retrieved all users")
	return users, nil
}

func (s *userService) DeleteUser(ctx context.Context, username string) error {
	if err := s.ValidateUsername(username); err != nil {
		return err
	}

	if err := s.userRepo.DeleteByUsername(ctx, username); err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.WithField("username", username).Info("User deleted successfully")
	return nil
}

func (s *userService) UserExists(ctx context.Context, username string) (bool, error) {
	if err := s.ValidateUsername(username); err != nil {
		return false, err
	}

	exists, err := s.userRepo.Exists(ctx, username)
	if err != nil {
		s.logger.WithError(err).WithField("username", username).Error("Failed to check if user exists")
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

func (s *userService) ValidateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) > s.config.Features.MaxUsernameLength {
		return fmt.Errorf("username too long (max %d characters)", s.config.Features.MaxUsernameLength)
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-' || char == ' ') {
			return fmt.Errorf("username contains invalid characters (only letters, numbers, spaces, underscore, and hyphen allowed)")
		}
	}

	// Check for reserved usernames
	reservedNames := []string{"admin", "system", "root", "moderator", "bot"}
	lowerUsername := strings.ToLower(strings.TrimSpace(username))
	for _, reserved := range reservedNames {
		if lowerUsername == reserved {
			return fmt.Errorf("username '%s' is reserved", username)
		}
	}

	return nil
}

func (s *userService) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total user count: %w", err)
	}

	onlineUsers, err := s.userRepo.GetOnlineUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}

	stats := map[string]interface{}{
		"total_users":         totalUsers,
		"online_users":        len(onlineUsers),
		"max_username_length": s.config.Features.MaxUsernameLength,
	}

	return stats, nil
}

func (s *userService) sanitizeUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.Join(strings.Fields(username), " ")

	return username
}
