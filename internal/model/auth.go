package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoginRequest struct {
	Username      string `json:"username" validate:"required,min=3,max=50"`
	Password      string `json:"password" validate:"required,min=6"`
	Captcha       string `json:"captcha" validate:"required"`
	CaptchaAnswer string `json:"captcha_answer" validate:"required"`
}

type RegisterRequest struct {
	Username      string `json:"username" validate:"required,min=3,max=50"`
	Email         string `json:"email" validate:"required,email"`
	Password      string `json:"password" validate:"required,min=6"`
	Age           int    `json:"age" validate:"min=13,max=150"`
	Gender        Gender `json:"gender" validate:"oneof=male female other private"`
	Bio           string `json:"bio" validate:"max=500"`
	Captcha       string `json:"captcha" validate:"required"`
	CaptchaAnswer string `json:"captcha_answer" validate:"required"`
}

type AuthResponse struct {
	Response
	User         interface{} `json:"user"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
	Captcha         string `json:"captcha" validate:"required"`
}

type ProfileUpdateRequest struct {
	Age      int    `json:"age" validate:"min=13,max=150"`
	Gender   Gender `json:"gender" validate:"oneof=male female other private"`
	Bio      string `json:"bio" validate:"max=500"`
}

type RefreshToken struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	Token      string             `json:"token" bson:"token"`
	ExpiresAt  time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	IsRevoked  bool               `json:"is_revoked" bson:"is_revoked"`
	DeviceInfo string             `json:"device_info,omitempty" bson:"device_info,omitempty"`
}

type Session struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Token     string             `json:"token" bson:"token"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	LastUsed  time.Time          `json:"last_used" bson:"last_used"`
	IPAddress string             `json:"ip_address" bson:"ip_address"`
	UserAgent string             `json:"user_agent" bson:"user_agent"`
	IsActive  bool               `json:"is_active" bson:"is_active"`
}

type CaptchaChallenge struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Challenge string             `json:"challenge" bson:"challenge"`
	Answer    string             `json:"-" bson:"answer"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	IsUsed    bool               `json:"is_used" bson:"is_used"`
	IPAddress string             `json:"ip_address" bson:"ip_address"`
}

func NewRefreshToken(userID primitive.ObjectID, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		IsRevoked: false,
	}
}

func NewSession(userID primitive.ObjectID, token string, expiresAt time.Time, ipAddress, userAgent string) *Session {
	now := time.Now()
	return &Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		LastUsed:  now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}
}

func NewCaptchaChallenge(challenge, answer string, ipAddress string) *CaptchaChallenge {
	return &CaptchaChallenge{
		ID:        primitive.NewObjectID(),
		Challenge: challenge,
		Answer:    answer,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5 minutes expiry
		CreatedAt: time.Now(),
		IsUsed:    false,
		IPAddress: ipAddress,
	}
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && !rt.IsExpired()
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

func (s *Session) UpdateLastUsed() {
	s.LastUsed = time.Now()
}

func (c *CaptchaChallenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *CaptchaChallenge) IsValid() bool {
	return !c.IsUsed && !c.IsExpired()
}

func (c *CaptchaChallenge) MarkAsUsed() {
	c.IsUsed = true
}
