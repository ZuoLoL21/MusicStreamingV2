package middleware

import (
	"context"
	"fmt"
	"libs/di"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ServiceJWTConfig provides configuration for service JWT middleware
type ServiceJWTConfig interface {
	GetUserUUIDKey() (any, bool)
	GetServiceJWTKey() (any, bool)
}

// ServiceJWTHandler generates service JWTs for authenticated requests
type ServiceJWTHandler struct {
	logger        *zap.Logger
	jwtHandler    *di.JWTHandler
	returns       *di.ReturnManager
	duration      time.Duration
	uuidKey       di.ContextKey
	serviceJWTKey di.ContextKey
}

// NewServiceJWTHandler creates a new service JWT middleware handler
func NewServiceJWTHandler(
	logger *zap.Logger,
	config ServiceJWTConfig,
	jwtHandler *di.JWTHandler,
	returns *di.ReturnManager,
	duration time.Duration,
) *ServiceJWTHandler {
	uuidKey, uuidOk := config.GetUserUUIDKey()
	if !uuidOk {
		logger.Error("not able to initialize ServiceJWTHandler: no user uuid key")
	}

	serviceJWTKey, serviceOk := config.GetServiceJWTKey()
	if !serviceOk {
		logger.Error("not able to initialize ServiceJWTHandler: no service jwt key")
	}

	return &ServiceJWTHandler{
		logger:        logger,
		jwtHandler:    jwtHandler,
		returns:       returns,
		duration:      duration,
		uuidKey:       uuidKey.(di.ContextKey),
		serviceJWTKey: serviceJWTKey.(di.ContextKey),
	}
}

// GetServiceJWTMiddleware returns middleware that generates service JWTs
// This middleware assumes the user JWT has already been validated by AuthHandler
// and the user UUID is available in the request context
func (h *ServiceJWTHandler) GetServiceJWTMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uuid, ok := r.Context().Value(h.uuidKey).(string)
			if !ok || uuid == "" {
				h.logger.Error("user UUID not found in context")
				h.returns.ReturnError(w, "internal server error: missing user context", http.StatusInternalServerError)
				return
			}

			// Generate service JWT using Vault Transit
			serviceJWT := h.jwtHandler.GenerateJwt(di.JWTSubjectService, uuid, h.duration)
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
			ctx := context.WithValue(r.Context(), h.serviceJWTKey, serviceJWT)
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
