package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"chatmix-backend/internal/model"
	"chatmix-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// UserHandler handles authentication and user-related requests
type UserHandler struct {
	authService service.AuthService
	userService service.UserService
	validator   *validator.Validate
	logger      *logrus.Logger
}

func NewUserHandler(
	authService service.AuthService,
	userService service.UserService,
	logger *logrus.Logger,
) *UserHandler {
	return &UserHandler{
		authService: authService,
		userService: userService,
		validator:   validator.New(),
		logger:      logger,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	ipAddress := h.getClientIP(r)

	authResponse, err := h.authService.Register(ctx, &req, ipAddress)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"username": req.Username,
			"email":    req.Email,
			"ip":       ipAddress,
		}).Error("Registration failed")

		switch {
		case strings.Contains(err.Error(), "already exists"):
			WriteError(w, http.StatusConflict, authResponse.Message)
		case authResponse.Code == 1:
			WriteError(w, http.StatusBadRequest, authResponse.Message)
		default:
			WriteError(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"username": req.Username,
		"email":    req.Email,
		"ip":       ipAddress,
	}).Info("User registered successfully")

	WriteJSON(w, http.StatusCreated, authResponse)
}

// Login handles user login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Get client info
	ipAddress := h.getClientIP(r)
	userAgent := r.UserAgent()

	// Login user
	authResponse, err := h.authService.Login(ctx, &req, ipAddress, userAgent)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"username": req.Username,
			"ip":       ipAddress,
		}).Error("Login failed")

		switch {
		case strings.Contains(err.Error(), "credentials"):
			WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		case strings.Contains(err.Error(), "captcha"):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	h.logger.WithFields(logrus.Fields{
		"username": req.Username,
		"ip":       ipAddress,
	}).Info("User logged in successfully")

	WriteJSON(w, http.StatusOK, authResponse)
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	authResponse, err := h.authService.RefreshToken(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Token refresh failed")
		WriteError(w, http.StatusUnauthorized, authResponse.Message)
		return
	}

	WriteJSON(w, http.StatusOK, authResponse)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	token := h.extractTokenFromHeader(r)
	if token == "" {
		WriteError(w, http.StatusBadRequest, "Token required")
		return
	}

	if err := h.authService.Logout(ctx, user.ID.Hex(), token); err != nil {
		h.logger.WithError(err).WithField("user_id", user.ID.Hex()).Error("Logout failed")
		WriteError(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req model.PasswordChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	if err := h.authService.ChangePassword(ctx, user.ID.Hex(), &req); err != nil {
		h.logger.WithError(err).WithField("user_id", user.ID.Hex()).Error("Password change failed")

		switch {
		case strings.Contains(err.Error(), "current password"):
			WriteError(w, http.StatusBadRequest, "Invalid current password")
		case strings.Contains(err.Error(), "captcha"):
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "Password change failed")
		}
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	WriteJSON(w, http.StatusOK, user.ToPrivateUser())
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req model.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Nickname field removed - using username only
	if req.Age > 0 {
		user.Age = req.Age
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	user.UpdatedAt = time.Now()

	if err := h.userService.UpdateUser(ctx, user); err != nil {
		h.logger.WithError(err).WithField("user_id", user.ID.Hex()).Error("Profile update failed")
		WriteError(w, http.StatusInternalServerError, "Profile update failed")
		return
	}

	WriteJSON(w, http.StatusOK, user.ToPrivateUser())
}

func (h *UserHandler) GenerateCaptcha(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ipAddress := h.getClientIP(r)

	challengeID, challenge, err := h.authService.GenerateCaptcha(ctx, ipAddress)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Captcha generation failed")
		return
	}

	response := map[string]string{
		"challenge_id": challengeID,
		"challenge":    challenge,
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *UserHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.authService.RevokeAllSessions(ctx, user.ID.Hex()); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to revoke sessions")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "All sessions revoked successfully",
	})
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	users, err := h.userService.GetAllUsers(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	publicUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		publicUsers[i] = user.ToPublicUser()
	}

	WriteJSON(w, http.StatusOK, publicUsers)
}

func (h *UserHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	users, err := h.userService.GetOnlineUsers(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get online users")
		return
	}

	publicUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		publicUsers[i] = user.ToPublicUser()
	}

	WriteJSON(w, http.StatusOK, publicUsers)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		WriteError(w, http.StatusBadRequest, "Username is required")
		return
	}

	user, err := h.userService.GetUser(ctx, username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	WriteJSON(w, http.StatusOK, user.ToPublicUser())
}

func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractTokenFromHeader(r)
		if token == "" {
			WriteError(w, http.StatusUnauthorized, "Authorization token required")
			return
		}

		user, err := h.authService.GetUserFromToken(token)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *UserHandler) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractTokenFromHeader(r)
		if token != "" {
			user, err := h.authService.GetUserFromToken(token)
			if err == nil {
				// Add user to context if token is valid
				ctx := context.WithValue(r.Context(), "user", user)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (h *UserHandler) extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func (h *UserHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
