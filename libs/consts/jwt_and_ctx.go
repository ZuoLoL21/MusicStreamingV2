package consts

import "time"

// ContextKey defines keys for context values.
// These keys are used to store and retrieve values from Go context.Context.
type ContextKey string

const (
	RequestIDKey  ContextKey = "requestId"
	UserUUIDKey   ContextKey = "userUuid"
	ServiceJWTKey ContextKey = "serviceJwt"
)

// JWT Subjects - these define the type of JWT token being created
const (
	JWTSubjectNormal  = "normal"
	JWTSubjectRefresh = "refresh"
	JWTSubjectService = "service"
)

// JWT Expiration Times - duration until token expiration
const (
	JWTExpirationNormal  = 10 * time.Minute
	JWTExpirationRefresh = 10 * 24 * time.Hour // 10 days
	JWTExpirationService = 2 * time.Minute
)

// JWT Header Keys used in JWT token headers
const (
	HeaderKeyID     = "kid"
	HeaderAppName   = "app_name"
	HeaderAlgorithm = "VaultSigningAlgorithm"
)

// Vault Operation Timeouts - timeouts for Vault signing/verification operations
const (
	JWTTimeoutVault = 30 * time.Second // Timeout for Vault Transit operations
)
