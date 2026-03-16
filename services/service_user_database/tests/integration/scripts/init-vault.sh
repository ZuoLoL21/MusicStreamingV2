#!/bin/sh
set -e

echo "Waiting for Vault..."
until vault status; do
  sleep 1
done

echo "Initializing Vault transit engine and keys..."
vault secrets enable -path=transit transit || echo "Transit already enabled"
vault write -f transit/keys/service-user-database type=rsa-4096 || echo "service-user-database key exists"
vault write -f transit/keys/gateway-api type=rsa-4096 || echo "gateway-api key exists"
vault write -f transit/keys/jwt-backend type=rsa-4096 || echo "jwt-backend key exists"
vault write -f transit/keys/jwt-user type=rsa-4096 || echo "jwt-user key exists"
vault write -f transit/keys/jwt-refresh type=rsa-4096 || echo "jwt-refresh key exists"

echo "Vault initialized successfully"
