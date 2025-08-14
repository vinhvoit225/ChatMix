package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"chatmix-backend/internal/config"
	"chatmix-backend/internal/model"
	"chatmix-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest, ipAddress string) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.AuthResponse, error)
	RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.AuthResponse, error)
	Logout(ctx context.Context, userID string, token string) error
	ValidateToken(tokenString string) (*jwt.Token, error)
	GetUserFromToken(tokenString string) (*model.User, error)
	ChangePassword(ctx context.Context, userID string, req *model.PasswordChangeRequest) error
	GenerateCaptcha(ctx context.Context, ipAddress string) (string, string, error)
	ValidateCaptcha(ctx context.Context, challenge, answer string) error
	RevokeAllSessions(ctx context.Context, userID string) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	sessionRepo      repository.SessionRepository
	captchaRepo      repository.CaptchaRepository
	config           *config.Config
	logger           *logrus.Logger
	jwtSecret        []byte
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	sessionRepo repository.SessionRepository,
	captchaRepo repository.CaptchaRepository,
	config *config.Config,
	logger *logrus.Logger,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		sessionRepo:      sessionRepo,
		captchaRepo:      captchaRepo,
		config:           config,
		logger:           logger,
		jwtSecret:        []byte(config.Auth.JWTSecret),
	}
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest, ipAddress string) (*model.AuthResponse, error) {
	response := &model.AuthResponse{}

	if err := s.ValidateCaptcha(ctx, req.Captcha, req.CaptchaAnswer); err != nil {
		response.Code = 1
		response.Message = "Invalid captcha"
		return response, err
	}

	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		response.Code = 2
		response.Message = "Failed to check user existence"
		return response, err
	}
	if existingUser != nil {
		response.Code = 3
		response.Message = "Username already exists"
		return response, err
	}

	existingEmail, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		response.Code = 4
		response.Message = "Failed to check email existence"
		return response, err
	}
	if existingEmail != nil {
		response.Code = 5
		response.Message = "Email already exists"
		return response, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Code = 6
		response.Message = "Failed to hash password"
		return response, err
	}

	user := model.NewUserWithProfile(req.Username, req.Email, req.Age, req.Gender, req.Bio)
	user.PasswordHash = string(hashedPassword)

	if !user.IsValid(s.config.Features.MaxUsernameLength) {
		response.Code = 7
		response.Message = "Invalid user data"
		return response, err
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		response.Code = 8
		response.Message = "Failed to create user"
		return response, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
	}).Info("User registered successfully")

	return s.generateTokensAndSession(ctx, user, ipAddress, "registration")
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.AuthResponse, error) {
	response := &model.AuthResponse{}
	if err := s.ValidateCaptcha(ctx, req.Captcha, req.CaptchaAnswer); err != nil {
		response.Code = 1
		response.Message = "Invalid captcha"
		return response, err
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		response.Code = 2
		response.Message = "Failed to get user"
		return response, err
	}

	if user == nil {
		user, err = s.userRepo.GetByEmail(ctx, req.Username)
		if err != nil {
			response.Code = 3
			response.Message = "Failed to get user"
			return response, err
		}
	}

	if user == nil {
		response.Code = 4
		response.Message = "Invalid credentials"
		return response, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Code = 5
		response.Message = "Invalid credentials"
		return response, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
	}).Info("User logged in successfully")

	return s.generateTokensAndSession(ctx, user, ipAddress, userAgent)
}

func (s *authService) generateTokensAndSession(ctx context.Context, user *model.User, ipAddress, userAgent string) (*model.AuthResponse, error) {
	response := &model.AuthResponse{}
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		response.Code = 1
		response.Message = "Failed to generate access token"
		return response, err
	}

	refreshTokenString, err := s.generateRandomToken()
	if err != nil {
		response.Code = 2
		response.Message = "Failed to generate refresh token"
		return response, err
	}

	refreshToken := model.NewRefreshToken(
		user.ID,
		refreshTokenString,
		time.Now().Add(time.Duration(s.config.Auth.RefreshTokenExpiry)*time.Hour),
	)
	refreshToken.DeviceInfo = userAgent

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		response.Code = 3
		response.Message = "Failed to save refresh token"
		return response, err
	}

	session := model.NewSession(user.ID, accessToken, expiresAt, ipAddress, userAgent)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		response.Code = 4
		response.Message = "Failed to create session"
		return response, err
	}

	if err := s.userRepo.SetOnlineStatus(ctx, user.Username, true); err != nil {
		response.Code = 5
		response.Message = "Failed to set user online"
		return response, err
	}

	return &model.AuthResponse{
		User:         user.ToPrivateUser(),
		Token:        accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *authService) generateAccessToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.config.Auth.AccessTokenExpiry) * time.Hour)

	claims := jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
		"iss":      "chatmix",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *authService) generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *authService) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.AuthResponse, error) {
	response := &model.AuthResponse{}
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, req.RefreshToken)
	if err != nil {
		response.Code = 1
		response.Message = "Invalid refresh token"
		return response, err
	}

	if refreshToken == nil || !refreshToken.IsValid() {
		response.Code = 2
		response.Message = "Invalid or expired refresh token"
		return response, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		response.Code = 3
		response.Message = "Failed to get user"
		return response, err
	}

	if user == nil {
		response.Code = 4
		response.Message = "User not found"
		return response, err
	}

	// Revoke old refresh token
	if err := s.refreshTokenRepo.Revoke(ctx, refreshToken.ID); err != nil {
		response.Code = 5
		response.Message = "Failed to revoke old refresh token"
		return response, err
	}

	return s.generateTokensAndSession(ctx, user, "", "token_refresh")
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
}

// GetUserFromToken extracts user from JWT token
func (s *authService) GetUserFromToken(tokenString string) (*model.User, error) {
	token, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := s.userRepo.GetByID(ctx, mustParseObjectID(userIDStr))
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Logout logs out a user and revokes session
func (s *authService) Logout(ctx context.Context, userID, token string) error {
	// Set user offline
	user, err := s.userRepo.GetByID(ctx, mustParseObjectID(userID))
	if err == nil && user != nil {
		if err := s.userRepo.SetOnlineStatus(ctx, user.Username, false); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to set user offline")
		}
	}

	// Deactivate session
	if err := s.sessionRepo.DeactivateByToken(ctx, token); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to deactivate session")
	}

	s.logger.WithField("user_id", userID).Info("User logged out")
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID string, req *model.PasswordChangeRequest) error {
	if err := s.ValidateCaptcha(ctx, req.Captcha, req.Captcha); err != nil {
		return fmt.Errorf("invalid captcha: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, mustParseObjectID(userID))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return fmt.Errorf("invalid current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all refresh tokens for security
	if err := s.refreshTokenRepo.RevokeAllByUserID(ctx, user.ID); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to revoke refresh tokens")
	}

	s.logger.WithField("user_id", userID).Info("Password changed successfully")
	return nil
}

// GenerateCaptcha generates a simple math captcha
func (s *authService) GenerateCaptcha(ctx context.Context, ipAddress string) (string, string, error) {
	// Generate simple math captcha
	a := randomInt(1, 20)
	b := randomInt(1, 20)
	operation := randomInt(0, 2) // 0: add, 1: subtract, 2: multiply

	var challenge, answer string
	switch operation {
	case 0:
		challenge = fmt.Sprintf("%d + %d = ?", a, b)
		answer = strconv.Itoa(a + b)
	case 1:
		if a < b {
			a, b = b, a // Ensure positive result
		}
		challenge = fmt.Sprintf("%d - %d = ?", a, b)
		answer = strconv.Itoa(a - b)
	case 2:
		challenge = fmt.Sprintf("%d × %d = ?", a, b)
		answer = strconv.Itoa(a * b)
	}

	// Create captcha record
	captcha := model.NewCaptchaChallenge(challenge, answer, ipAddress)
	if err := s.captchaRepo.Create(ctx, captcha); err != nil {
		return "", "", fmt.Errorf("failed to create captcha: %w", err)
	}

	return captcha.ID.Hex(), challenge, nil
}

// ValidateCaptcha validates a captcha answer
func (s *authService) ValidateCaptcha(ctx context.Context, challengeID, answer string) error {
	captcha, err := s.captchaRepo.GetByID(ctx, mustParseObjectID(challengeID))
	if err != nil {
		return fmt.Errorf("invalid captcha")
	}

	if captcha == nil || !captcha.IsValid() {
		return fmt.Errorf("captcha expired or already used")
	}

	if captcha.Answer != answer {
		return fmt.Errorf("incorrect captcha answer")
	}

	captcha.MarkAsUsed()
	if err := s.captchaRepo.Update(ctx, captcha); err != nil {
		s.logger.WithError(err).Error("Failed to mark captcha as used")
	}

	return nil
}

// RevokeAllSessions revokes all sessions for a user
func (s *authService) RevokeAllSessions(ctx context.Context, userID string) error {
	userOID := mustParseObjectID(userID)

	if err := s.sessionRepo.DeactivateAllByUserID(ctx, userOID); err != nil {
		return fmt.Errorf("failed to deactivate sessions: %w", err)
	}

	if err := s.refreshTokenRepo.RevokeAllByUserID(ctx, userOID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

func randomInt(min, max int) int {
	b := make([]byte, 1)
	rand.Read(b)
	return min + int(b[0])%(max-min+1)
}

func mustParseObjectID(id string) primitive.ObjectID {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(fmt.Sprintf("invalid ObjectID: %s", id))
	}
	return oid
}
