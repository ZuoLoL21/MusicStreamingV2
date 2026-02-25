package middleware

import (
	"context"
	"fmt"
	"libs/di"
	"libs/helpers"
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
	secrets       *di.SecretsManager
	returns       *di.ReturnManager
	jwtStorePath  string
	duration      time.Duration
	uuidKey       di.ContextKey
	serviceJWTKey di.ContextKey
}

// NewServiceJWTHandler creates a new service JWT middleware handler
func NewServiceJWTHandler(
	logger *zap.Logger,
	config ServiceJWTConfig,
	secrets *di.SecretsManager,
	returns *di.ReturnManager,
	jwtStorePath string,
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
		secrets:       secrets,
		returns:       returns,
		jwtStorePath:  jwtStorePath,
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

			_, privateKey, kid := h.secrets.GetKeyInfo(h.jwtStorePath)

			// Add service JWT to context
			serviceJWT := helpers.GenerateServiceJwt(uuid, privateKey, kid, h.duration)
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
