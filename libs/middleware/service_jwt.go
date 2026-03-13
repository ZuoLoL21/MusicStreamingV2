package middleware

import (
	"context"
	"fmt"
	"libs/consts"
	"libs/di"
	"libs/helpers"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ServiceJWTHandler handles service-to-service JWT generation.
// It generates short-lived JWT tokens for authenticated service-to-service communication.
type ServiceJWTHandler struct {
	logger     *zap.Logger
	jwtHandler *di.JWTHandler
	returns    *di.ReturnManager
	duration   time.Duration
}

// NewServiceJWTHandler creates a new ServiceJWTHandler for generating service JWTs.
//
// Parameters:
//   - logger: Zap logger for logging token generation events
//   - jwtHandler: JWT handler for token generation
//   - returns: ReturnManager for writing responses
//   - duration: Token validity duration (typically short for service tokens)
//
// Returns a configured ServiceJWTHandler ready for middleware creation.
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

// GetServiceJWTMiddleware returns a Gorilla Mux middleware that generates service JWTs.
//
// The middleware extracts the user UUID from the context (set by auth middleware),
// generates a short-lived service JWT with that UUID, and stores it in the context.
// This JWT can be used for service-to-service authentication.
//
// Returns 500 Internal Server Error if user UUID is missing or token generation fails.
func (h *ServiceJWTHandler) GetServiceJWTMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uuid := helpers.GetUserUUIDFromContext(r.Context())
			if uuid == "" {
				h.logger.Error("user UUID not found in context")
				h.returns.ReturnError(w, "internal server error: missing user context", http.StatusInternalServerError)
				return
			}

			serviceJWT, err := h.jwtHandler.GenerateJwt(consts.JWTSubjectService, uuid, h.duration)
			if err != nil {
				h.logger.Error("failed to generate service JWT",
					zap.String("user_uuid", uuid),
					zap.Error(err))
				h.returns.ReturnError(w, "internal server error: failed to generate service token", http.StatusInternalServerError)
				return
			}

			h.logger.Info("service JWT generated",
				zap.String("user_uuid", uuid),
				zap.Duration("ttl", h.duration))

			ctx := context.WithValue(r.Context(), consts.ServiceJWTKey, serviceJWT)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractServiceJWT extracts a service JWT from the context using a custom key.
// The key parameter allows specifying a custom context key for extraction.
// Returns the JWT string if found, or an error if not present in context.
// This is useful for retrieving the service JWT set by GetServiceJWTMiddleware.
func ExtractServiceJWT(ctx context.Context, key any) (string, error) {
	jwt, ok := ctx.Value(key).(string)
	if !ok || jwt == "" {
		return "", fmt.Errorf("service JWT not found in context")
	}
	return jwt, nil
}
