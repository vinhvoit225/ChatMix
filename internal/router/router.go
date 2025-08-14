package router

import (
	"net/http"

	"chatmix-backend/internal/config"
	"chatmix-backend/internal/handler"
	"chatmix-backend/internal/service"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Router manages HTTP routes
type Router struct {
	mux         *mux.Router
	config      *config.Config
	logger      *logrus.Logger
	httpHandler *handler.HTTPHandler
	authHandler *handler.UserHandler
	chatHandler *handler.ChatHandler
}

func NewRouter(
	config *config.Config,
	logger *logrus.Logger,
	httpHandler *handler.HTTPHandler,
	authHandler *handler.UserHandler,
	authService service.AuthService,
	chatHandler *handler.ChatHandler,
) *Router {

	return &Router{
		mux:         mux.NewRouter(),
		config:      config,
		logger:      logger,
		httpHandler: httpHandler,
		authHandler: authHandler,
		chatHandler: chatHandler,
	}
}

func (r *Router) SetupRoutes() *mux.Router {
	r.mux.Use(r.httpHandler.RecoveryMiddleware)
	r.mux.Use(r.httpHandler.LoggingMiddleware)
	r.mux.Use(r.httpHandler.CORSMiddleware(r.config.Server.CORS))
	r.mux.Methods("OPTIONS").HandlerFunc(r.handleOptions)

	// API routes
	api := r.mux.PathPrefix("/api").Subrouter()
	r.setupAPIRoutes(api)

	// WebSocket chat route (handles auth internally via token query param)
	r.mux.HandleFunc("/ws/chat", r.chatHandler.HandleWebSocket).Methods("GET")

	// Health check
	r.mux.HandleFunc("/health", r.httpHandler.HealthCheck).Methods("GET")

	return r.mux
}

func (r *Router) setupAPIRoutes(api *mux.Router) {
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", r.authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", r.authHandler.Login).Methods("POST")
	auth.HandleFunc("/refresh", r.authHandler.RefreshToken).Methods("POST")
	auth.HandleFunc("/captcha", r.authHandler.GenerateCaptcha).Methods("GET")

	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(r.authHandler.AuthMiddleware)
	authProtected.HandleFunc("/logout", r.authHandler.Logout).Methods("POST")
	authProtected.HandleFunc("/change-password", r.authHandler.ChangePassword).Methods("POST")
	authProtected.HandleFunc("/profile", r.authHandler.GetProfile).Methods("GET")
	authProtected.HandleFunc("/profile", r.authHandler.UpdateProfile).Methods("PUT")
	authProtected.HandleFunc("/revoke-sessions", r.authHandler.RevokeAllSessions).Methods("POST")

	chatProtected := api.PathPrefix("/chat").Subrouter()
	chatProtected.Use(r.authHandler.AuthMiddleware)
	chatProtected.HandleFunc("/start", r.chatHandler.HandleStartChat).Methods("POST")
	chatProtected.HandleFunc("/queue-status", r.chatHandler.HandleQueueStatus).Methods("GET")

	api.HandleFunc("/users", r.authHandler.GetUsers).Methods("GET")
	api.HandleFunc("/users/online", r.authHandler.GetOnlineUsers).Methods("GET")
	api.HandleFunc("/users/{username}", r.authHandler.GetUser).Methods("GET")

	api.HandleFunc("/health", r.httpHandler.HealthCheck).Methods("GET")
}

func (r *Router) ListRoutes() []string {
	var routes []string

	err := r.mux.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		methods, err := route.GetMethods()
		if err != nil {
			routes = append(routes, path)
		} else {
			for _, method := range methods {
				routes = append(routes, method+" "+path)
			}
		}

		return nil
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to list routes")
	}

	return routes
}

// handleOptions handles OPTIONS requests for CORS preflight
func (r *Router) handleOptions(w http.ResponseWriter, req *http.Request) {
	// CORS headers are already set by middleware
	// Just return 200 OK
	w.WriteHeader(http.StatusOK)
}
