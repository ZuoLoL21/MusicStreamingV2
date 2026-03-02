module gateway_api

go 1.24.3

require (
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.1
	libs v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/vault-client-go v0.4.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/time v0.12.0 // indirect
)

replace libs => ../../libs
