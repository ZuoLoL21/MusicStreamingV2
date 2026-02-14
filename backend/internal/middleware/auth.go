package middleware

import (
	"backend/internal/dependencies"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var allowedRoutes = []RoutePattern{
	{BuildRegex("login"), "POST"},
	{BuildRegex("login"), "PUT"},
}

var refreshRoutes = []RoutePattern{
	{BuildRegex("renew"), "POST"},
}

type AuthHandler struct {
	logger *zap.Logger
	config *dependencies.Config
}

func NewAuthHandler(logger *zap.Logger, config *dependencies.Config) *AuthHandler {
	return &AuthHandler{logger: logger, config: config}
}

func (h *AuthHandler) GetAuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentRoute := Route{r.URL.Path, r.Method}

			h.logger.Info("incoming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			for _, route := range allowedRoutes {
				if route.Matches(currentRoute) {
					next.ServeHTTP(w, r)
					return
				}
			}

			var authorized bool
			var err error
			var matched bool

			for _, route := range refreshRoutes {
				if route.Matches(currentRoute) {
					matched = true
					authorized, err = h.handleSpecialAuth(w, r)
				}
			}
			if !matched {
				authorized, err = h.handleDefaultAuth(w, r)
			}

			if err != nil {
				h.logger.Warn("auth failed completely",
					zap.Error(err),
					zap.String("route", r.URL.Path),
					zap.String("method", r.Method),
				)
			}
			if authorized {
				http.Error(w, "unauthorized", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
			return
		})
	}
}

func (h *AuthHandler) handleDefaultAuth(w http.ResponseWriter, r *http.Request) (bool, error) {

}

func (h *AuthHandler) handleSpecialAuth(w http.ResponseWriter, r *http.Request) (bool, error) {

}
