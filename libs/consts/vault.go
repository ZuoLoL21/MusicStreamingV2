package consts

// Vault Configuration
const (
	VaultMountPath     = "transit"
	VaultHashAlgorithm = "sha2-256"
	VaultMarshalingAlg = "jws"
)

// Vault Return Indices
const (
	VaultSignatureKey = "signature"
	VaultValidKey     = "valid"
)

// Vault Parsing

const (
	VaultSignaturePrefix = "vault"
	VaultSignatureFormat = "vault:v%d:%s"
)
