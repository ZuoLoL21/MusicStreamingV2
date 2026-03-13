package consts

// Package consts contains shared constants for the music streaming application.
// This file defines Vault Transit application names used for JWT signing keys.

// Vault Transit application names for each service.
// These correspond to the Transit keys used for JWT signing in HashiCorp Vault.
// Each service in the system has its own dedicated transit key in Vault.
const (
	VaultAppGatewayAPI            = "gateway_api"
	VaultAppUserDatabase          = "service_user_database"
	VaultAppGatewayRecommendation = "gateway_recommendation"
	VaultAppPopularitySystem      = "service_popularity_system"
	VaultAppEventIngestion        = "service_event_ingestion"
)
