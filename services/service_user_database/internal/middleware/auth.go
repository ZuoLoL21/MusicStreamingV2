package middleware

import (
	"backend/internal/di"
	"context"
	"fmt"
	di2 "libs/di"
	"net/http"
	"strings"

	libshelpers "libs/helpers"

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

type AuthHandler struct {
	logger  *zap.Logger
	config  *di.Config
	secrets *di2.SecretsManager
}

func NewAuthHandler(logger *zap.Logger, config *di.Config, secrets *di2.SecretsManager) *AuthHandler {
	return &AuthHandler{logger: logger, config: config, secrets: secrets}
}

func (h *AuthHandler) GetAuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			template, _ := mux.CurrentRoute(r).GetPathTemplate()
			currentRoute := Route{template, r.Method}

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

			ctx := context.WithValue(r.Context(), h.config.UserUUIDKey, uuid)
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

	uuid, err := libshelpers.ValidateJwt(subject, token, h.secrets.GetPublicKeyFunc())
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
