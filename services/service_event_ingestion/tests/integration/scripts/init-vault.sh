#!/bin/sh
set -e

echo "Waiting for Vault..."
until vault status; do
  sleep 1
done

echo "Initializing Vault transit engine..."
vault secrets enable -path=transit transit || echo "Transit already enabled"
vault write -f transit/keys/event-ingestion type=rsa-4096 || echo "event-ingestion key exists"

echo "Vault initialized successfully"
