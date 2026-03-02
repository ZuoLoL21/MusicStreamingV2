package middleware

import (
	"context"
	"fmt"
	"libs/di"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type AuthConfig interface {
	GetUserUUIDKey() (any, bool) // Returns (key, hasUserUUID)
}

type AuthHandler struct {
	logger     *zap.Logger
	jwtHandler *di.JWTHandler
	returns    *di.ReturnManager
	subject    string
	uuidKey    di.ContextKey
}

func NewAuthHandler(
	logger *zap.Logger,
	config AuthConfig,
	jwtHandler *di.JWTHandler,
	returns *di.ReturnManager,
	subject string,
) *AuthHandler {
	uuidKey, ok := config.GetUserUUIDKey()
	if !ok {
		logger.Error("not able to initialize AuthHandler: no user uuid key")
	}
	return &AuthHandler{
		logger:     logger,
		jwtHandler: jwtHandler,
		returns:    returns,
		subject:    subject,
		uuidKey:    uuidKey.(di.ContextKey),
	}
}

func (h *AuthHandler) GetAuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := h.parseToken(r)

			if err != nil {
				h.returns.ReturnError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			uuid, err := h.authenticate(token, h.subject)

			if err != nil {
				h.returns.ReturnError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), h.uuidKey, uuid)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		})
	}
}

func (h *AuthHandler) parseToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header")
	}
	return tokenParts[1], nil
}

func (h *AuthHandler) authenticate(token string, subject string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("invalid jwt")
	}

	uuid, err := h.jwtHandler.ValidateJwt(subject, token)
	if err != nil {
		if strings.Contains(err.Error(), "transit") || strings.Contains(err.Error(), "vault") {
			h.logger.Error("vault transit verify failed",
				zap.String("operation", "ValidateJwt"),
				zap.Error(err))
		} else {
			h.logger.Warn("authentication failed",
				zap.String("subject", subject),
				zap.String("reason", "invalid_token"),
				zap.Error(err))
		}
		return "", fmt.Errorf("invalid jwt: %v", err.Error())
	}

	h.logger.Info("authentication successful",
		zap.String("subject", subject),
		zap.String("user_uuid", uuid))

	return uuid, nil
}
