package handler

import (
	"net/http"
	"strings"
	"time"

	"chatmix-backend/internal/config"
	"chatmix-backend/internal/service"

	"github.com/sirupsen/logrus"
)

type HTTPHandler struct {
	userService service.UserService
	logger      *logrus.Logger
}

func NewHTTPHandler(
	userService service.UserService,
	logger *logrus.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":         "healthy",
		"timestamp":      time.Now(),
		"active_clients": 0,
	}

	WriteJSON(w, http.StatusOK, health)
}

func (h *HTTPHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := NewStatusResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		h.logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.String(),
			"status":      wrapped.Status(),
			"duration":    time.Since(start),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("HTTP request")
	})
}

func (h *HTTPHandler) CORSMiddleware(corsConfig config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range corsConfig.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			// Set CORS headers - always set for OPTIONS requests
			if allowed || r.Method == "OPTIONS" {
				if len(corsConfig.AllowedOrigins) == 1 && corsConfig.AllowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			// Always set CORS headers for allowed origins or OPTIONS requests
			if allowed || r.Method == "OPTIONS" {
				if len(corsConfig.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
				}

				if len(corsConfig.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
				}

				if corsConfig.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				WriteStatus(w, http.StatusOK)
				return
			}

			// Block request if origin not allowed (except for requests without Origin header)
			if origin != "" && !allowed {
				WriteStatus(w, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Recovery middleware
func (h *HTTPHandler) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				h.logger.WithField("error", err).Error("Panic recovered")
				WriteError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
