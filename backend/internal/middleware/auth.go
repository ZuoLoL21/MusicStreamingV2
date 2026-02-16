package middleware

import (
	"backend/internal/dependencies"
	"backend/internal/helpers"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Route struct {
	Route  string
	Method string
}

var allowedRoutes = []Route{
	{"/login", "POST"},
	{"/login", "PUT"},
}

var refreshRoutes = []Route{
	{"/renew", "POST"},
}

const UserUUIDKey = "uuidKey"

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
			template, _ := mux.CurrentRoute(r).GetPathTemplate()
			currentRoute := Route{template, r.Method}

			h.logger.Info("incoming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			for _, route := range allowedRoutes {
				if route == currentRoute {
					next.ServeHTTP(w, r)
					return
				}
			}

			var err error
			var matched bool
			var uuid string

			for _, route := range refreshRoutes {
				if route == currentRoute {
					matched = true
					uuid, err = h.handleRefreshAuth(r)
				}
			}
			if !matched {
				uuid, err = h.handleDefaultAuth(r)
			}

			if err != nil {
				h.logger.Warn("invalid token",
					zap.Error(err),
					zap.String("route", r.URL.Path),
					zap.String("method", r.Method),
				)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserUUIDKey, uuid)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		})
	}
}

func (h *AuthHandler) parseToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer ")
	}
	return ""
}

func (h *AuthHandler) authenticate(r *http.Request, subject string) (string, error) {
	token := h.parseToken(r)
	if token == "" {
		return "", fmt.Errorf("invalid jwt: missing \"Bearer \"")
	}

	uuid, err := helpers.ValidateJwt(subject, token, &h.config.PublicKey)
	if err != nil {
		h.logger.Info("auth failed completely", zap.Error(err))
		return "", fmt.Errorf("invalid jwt: %v", err.Error())
	}
	return uuid, nil
}

func (h *AuthHandler) handleDefaultAuth(r *http.Request) (string, error) {
	return h.authenticate(r, h.config.SubjectNormal)
}

func (h *AuthHandler) handleRefreshAuth(r *http.Request) (string, error) {
	return h.authenticate(r, h.config.SubjectRefresh)
}
