package middleware

import (
	"context"
	"fmt"
	"libs/consts"
	"libs/di"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AuthHandler handles JWT-based authentication for HTTP requests.
//
// It validates Bearer tokens from the Authorization header and extracts
// the user UUID to be stored in the request context.
type AuthHandler struct {
	logger     *zap.Logger
	jwtHandler *di.JWTHandler
	returns    *di.ReturnManager
	subject    string
}

// NewAuthHandler creates a new AuthHandler with the given dependencies.
//
// Parameters:
//   - logger: Zap logger for logging authentication events
//   - jwtHandler: JWT handler for token validation
//   - returns: ReturnManager for writing responses
//   - subject: Expected JWT subject (e.g., "normal" for user tokens)
//
// Returns a configured AuthHandler ready for middleware creation.
func NewAuthHandler(
	logger *zap.Logger,
	jwtHandler *di.JWTHandler,
	returns *di.ReturnManager,
	subject string,
) *AuthHandler {
	return &AuthHandler{
		logger:     logger,
		jwtHandler: jwtHandler,
		returns:    returns,
		subject:    subject,
	}
}

// GetAuthMiddleware returns a Gorilla Mux middleware function that performs JWT authentication.
// The middleware extracts the Bearer token from the Authorization header, validates it
// using the JWT handler, and adds the user UUID to the request context on success.
//
// Returns 401 Unauthorized if the token is missing, invalid, or has an unexpected subject.
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

			ctx := context.WithValue(r.Context(), consts.UserUUIDKey, uuid)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		})
	}
}

// parseToken extracts the JWT token from the Authorization header.
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

// authenticate validates the JWT token and returns the user's UUID if valid.
// It uses the JWT handler to validate the token against the expected subject.
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
