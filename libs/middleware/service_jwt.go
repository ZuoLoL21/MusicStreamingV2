package middleware

import (
	"context"
	"fmt"
	"libs/consts"
	"libs/di"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ServiceJWTHandler generates service JWTs for authenticated requests
type ServiceJWTHandler struct {
	logger     *zap.Logger
	jwtHandler *di.JWTHandler
	returns    *di.ReturnManager
	duration   time.Duration
}

// NewServiceJWTHandler creates a new service JWT middleware handler
func NewServiceJWTHandler(
	logger *zap.Logger,
	jwtHandler *di.JWTHandler,
	returns *di.ReturnManager,
	duration time.Duration,
) *ServiceJWTHandler {
	return &ServiceJWTHandler{
		logger:     logger,
		jwtHandler: jwtHandler,
		returns:    returns,
		duration:   duration,
	}
}

// GetServiceJWTMiddleware returns middleware that generates service JWTs
// This middleware assumes the user JWT has already been validated by AuthHandler
// and the user UUID is available in the request context
func (h *ServiceJWTHandler) GetServiceJWTMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uuid, ok := r.Context().Value(consts.UserUUIDKey).(string)
			if !ok || uuid == "" {
				h.logger.Error("user UUID not found in context")
				h.returns.ReturnError(w, "internal server error: missing user context", http.StatusInternalServerError)
				return
			}

			// Generate service JWT using Vault Transit
			serviceJWT := h.jwtHandler.GenerateJwt(consts.JWTSubjectService, uuid, h.duration)
			if serviceJWT == "" {
				h.logger.Error("failed to generate service JWT",
					zap.String("user_uuid", uuid))
				h.returns.ReturnError(w, "internal server error: failed to generate service token", http.StatusInternalServerError)
				return
			}

			h.logger.Info("service JWT generated",
				zap.String("user_uuid", uuid),
				zap.Duration("ttl", h.duration))

			// Add service JWT to context
			ctx := context.WithValue(r.Context(), consts.ServiceJWTKey, serviceJWT)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractServiceJWT is a helper to extract service JWT from context
func ExtractServiceJWT(ctx context.Context, key any) (string, error) {
	jwt, ok := ctx.Value(key).(string)
	if !ok || jwt == "" {
		return "", fmt.Errorf("service JWT not found in context")
	}
	return jwt, nil
}
