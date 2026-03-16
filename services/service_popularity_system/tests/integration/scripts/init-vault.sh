#!/bin/sh
set -e

echo "Waiting for Vault..."
until vault status; do
  sleep 1
done

echo "Initializing Vault transit engine..."
vault secrets enable -path=transit transit || echo "Transit already enabled"
vault write -f transit/keys/popularity-system type=rsa-4096 || echo "popularity-system key exists"

echo "Vault initialized successfully"
