package consts

// Vault Transit application names for each service
// These correspond to the Transit keys used for JWT signing
const (
	VaultAppGatewayAPI            = "gateway_api"
	VaultAppUserDatabase          = "service_user_database"
	VaultAppGatewayRecommendation = "gateway_recommendation"
	VaultAppPopularitySystem      = "service_popularity_system"
)
