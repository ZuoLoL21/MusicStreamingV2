package handlers

import (
	"encoding/json"
	"gateway_recommendation/internal/clients"
	libshandlers "libs/helpers"
	"net/http"

	libsdi "libs/di"

	"go.uber.org/zap"
)

type RecommendHandler struct {
	banditClient     *clients.BanditClient
	popularityClient *clients.PopularityClient
	returnManager    *libsdi.ReturnManager
	logger           *zap.Logger
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
	logger *zap.Logger,
) *RecommendHandler {
	return &RecommendHandler{
		banditClient:     banditClient,
		popularityClient: popularityClient,
		returnManager:    returnManager,
		logger:           logger,
	}
}

func (h *RecommendHandler) RecommendTheme(w http.ResponseWriter, r *http.Request) {
	requestID := libshandlers.GetRequestIDFromContext(r.Context())

	// Parse request body
	var req RecommendThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode request", zap.String("request_id", requestID), zap.Error(err))
		h.returnManager.ReturnError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserUUID == "" {
		h.logger.Warn("Missing user_uuid in request", zap.String("request_id", requestID))
		h.returnManager.ReturnError(w, "user_uuid is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing theme recommendation request",
		zap.String("request_id", requestID),
		zap.String("user_uuid", req.UserUUID))

	// Call bandit service for personalized prediction
	predictResp, err := h.banditClient.Predict(r.Context(), req.UserUUID, requestID)
	if err != nil {
		h.logger.Warn("Bandit prediction failed", zap.String("request_id", requestID), zap.String("user_uuid", req.UserUUID), zap.Error(err))
		h.returnManager.ReturnError(w, "Failed to get recommendation from bandit service", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Bandit prediction completed", zap.String("request_id", requestID), zap.String("recommended_theme", predictResp.Theme))

	// Extract service JWT
	serviceJWT := libshandlers.GetServiceJWTFromContext(r.Context())
	if serviceJWT == "" {
		h.logger.Error("Service JWT not found in Authorization header", zap.String("request_id", requestID))
		h.returnManager.ReturnError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch popularity data for themes
	popularityData, err := h.popularityClient.GetThemePopularity(r.Context(), requestID, serviceJWT, 10)
	if err != nil {
		h.logger.Warn("Failed to fetch popularity data", zap.String("request_id", requestID), zap.Error(err))
		h.returnManager.ReturnError(w, "Failed to fetch popularity data", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Popularity data fetched",
		zap.String("request_id", requestID),
		zap.Int("theme_count", len(popularityData)))

	// Combine and return enriched response
	response := RecommendThemeResponse{
		RecommendedTheme: predictResp.Theme,
		ThemeFeatures:    predictResp.Features,
		PopularityData:   popularityData,
	}

	h.logger.Info("Theme recommendation completed successfully",
		zap.String("request_id", requestID),
		zap.String("recommended_theme", response.RecommendedTheme))

	h.returnManager.ReturnJSON(w, response, http.StatusOK)
}
