package vault

const (
	// JWT Header Keys
	HeaderKeyID     = "kid"
	HeaderAppName   = "app_name"
	HeaderAlgorithm = "VaultSigningAlgorithm"

	// Vault Configuration
	VaultMountPath     = "transit"
	VaultHashAlgorithm = "sha2-256"
	VaultMarshalingAlg = "jws"

	// Vault Return Indices
	VaultSignatureKey = "signature"
	VaultValidKey     = "valid"

	// Vault Parsing
	VaultSignaturePrefix = "vault"
	VaultSignatureFormat = "vault:v%d:%s"

	// Error Messages
	ErrKIDMissing        = "kid header missing or invalid"
	ErrKIDNotInt         = "kid header not a proper int %x"
	ErrAppNameMissing    = "app_name header missing or invalid"
	ErrInvalidSubject    = "invalid subject - required %v, current %v"
	ErrInvalidToken      = "invalid token"
	ErrInvalidTransitKey = "invalid transit key"
	ErrInvalidFormat     = "invalid format"
	ErrInvalidKey        = "invalid key: must be of type *SigningKey"
	ErrVaultAddrMissing  = "VAULT_ADDR environment variable not set"
	ErrVaultTokenMissing = "VAULT_TOKEN environment variable not set"
)
