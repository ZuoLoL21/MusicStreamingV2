package consts

// Env Error Messages
const (
	ErrVaultAddrMissing  = "VAULT_ADDR environment variable not set"
	ErrVaultTokenMissing = "VAULT_TOKEN environment variable not set"
)

// JWT Error Messages
const (
	ErrKIDMissing        = "kid header missing or invalid"
	ErrKIDNotInt         = "kid header not a proper int %x"
	ErrAppNameMissing    = "app_name header missing or invalid"
	ErrInvalidSubject    = "invalid subject - required %v, current %v"
	ErrInvalidToken      = "invalid token"
	ErrInvalidTransitKey = "invalid transit key"
	ErrInvalidFormat     = "invalid format"
	ErrInvalidKey        = "invalid key: must be of type *SigningKey"
	ErrMissingUuid       = "uuid missing from token claims"
	ErrMissingDeviceID   = "device_id missing from token claims"
)

// HTTP Client Error Messages
const (
	ErrMarshalRequest    = "marshal request"
	ErrCreateRequest     = "create request"
	ErrRequestFailed     = "request failed"
	ErrReadResponse      = "read response"
	ErrUnmarshalResponse = "unmarshal response"
	ErrBuildURL          = "build URL"
	ErrParseBaseURL      = "parse base URL"
)

// Authentication Error Messages
const (
	ErrMissingAuthHeader     = "missing authorization header"
	ErrInvalidAuthHeader     = "invalid authorization header"
	ErrInvalidJWT            = "invalid jwt"
	ErrMissingUserContext    = "internal server error: missing user context"
	ErrFailedToGenerateToken = "internal server error: failed to generate service token"
	ErrServiceJWTNotFound    = "service JWT not found in context"
)

// HTTP Handler Error Messages
const (
	ErrUnauthorized        = "unauthorized"
	ErrInternalServerError = "internal server error"
	ErrBadGateway          = "bad gateway"
)

// Vault Error Messages
const (
	ErrVaultNotConfigured    = "vault address or token not configured"
	ErrKeyVersionFetchFailed = "failed after %d attempts"
	ErrHTTPRequestFailed     = "HTTP request failed"
	ErrVaultReturnedError    = "vault returned %d: %s"
	ErrDecodeResponse        = "failed to decode response"
	ErrLatestVersionMissing  = "latest_version is 0 or missing"
	ErrMarshalRequestFailed  = "failed to marshal request"
	ErrSignatureNotFound     = "signature not found in response"
	ErrParseVersionFailed    = "failed to parse version"
	ErrValidFieldNotFound    = "valid field not found in response"
)
