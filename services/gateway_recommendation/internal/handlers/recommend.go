package handlers

import (
	"gateway_recommendation/internal/clients"
	libshelpers "libs/helpers"
	libsmiddleware "libs/middleware"
	"net/http"

	libsdi "libs/di"

	"go.uber.org/zap"
)

type RecommendHandler struct {
	banditClient     *clients.BanditClient
	popularityClient *clients.PopularityClient
	returnManager    *libsdi.ReturnManager
}

type RecommendThemeRequest struct {
	UserUUID string `json:"user_uuid"`
}

type RecommendThemeResponse struct {
	RecommendedTheme string                    `json:"recommended_theme"`
	ThemeFeatures    []float64                 `json:"theme_features"`
	PopularityData   []clients.ThemePopularity `json:"popularity_data"`
}

func NewRecommendHandler(
	banditClient *clients.BanditClient,
	popularityClient *clients.PopularityClient,
	returnManager *libsdi.ReturnManager,
) *RecommendHandler {
	return &RecommendHandler{
		banditClient:     banditClient,
		popularityClient: popularityClient,
		returnManager:    returnManager,
	}
}

func (h *RecommendHandler) RecommendTheme(w http.ResponseWriter, r *http.Request) {
	logger := libsmiddleware.GetLogger(r.Context())

	requestID := libshelpers.GetRequestIDFromContext(r.Context())

	// Parse
	userUUID := libshelpers.GetUserUUIDFromContext(r.Context())
	if userUUID == "" {
		logger.Warn("Missing user_uuid in context")
		h.returnManager.ReturnError(w, "user_uuid not found in request context", http.StatusUnauthorized)
		return
	}
	logger.Debug("Processing theme recommendation request", zap.String("user_uuid", userUUID))

	// Call bandit service for personalized prediction
	predictResp, err := h.banditClient.Predict(r.Context(), userUUID, requestID)
	if err != nil {
		logger.Error("Bandit prediction failed", zap.Error(err))
		h.returnManager.ReturnError(w, "Failed to get recommendation", http.StatusServiceUnavailable)
		return
	}
	logger.Debug("Bandit prediction completed", zap.String("recommended_theme", predictResp.Theme))

	// Extract service JWT
	serviceJWT := libshelpers.GetServiceJWTFromContext(r.Context())
	if serviceJWT == "" {
		logger.Error("Service JWT not found in Authorization header")
		h.returnManager.ReturnError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch popularity data for themes
	popularityData, err := h.popularityClient.GetThemePopularity(r.Context(), requestID, serviceJWT, 10)
	if err != nil {
		logger.Warn("Failed to fetch popularity data", zap.Error(err))
		h.returnManager.ReturnError(w, "Failed to fetch popularity data", http.StatusInternalServerError)
		return
	}

	logger.Debug("Popularity data fetched", zap.Int("theme_count", len(popularityData)))

	// Combine and return enriched response
	response := RecommendThemeResponse{
		RecommendedTheme: predictResp.Theme,
		ThemeFeatures:    predictResp.Features,
		PopularityData:   popularityData,
	}

	logger.Info("Theme recommendation completed successfully",
		zap.String("recommended_theme", response.RecommendedTheme))

	h.returnManager.ReturnJSON(w, response, http.StatusOK)
}
