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

type ServiceJWTHandler struct {
	logger     *zap.Logger
	jwtHandler *di.JWTHandler
	returns    *di.ReturnManager
	duration   time.Duration
}

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

func ExtractServiceJWT(ctx context.Context, key any) (string, error) {
	jwt, ok := ctx.Value(key).(string)
	if !ok || jwt == "" {
		return "", fmt.Errorf("service JWT not found in context")
	}
	return jwt, nil
}
