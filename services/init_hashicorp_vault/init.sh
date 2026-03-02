#!/bin/sh
set -e

VAULT_ADDR="${VAULT_ADDR:-http://vault:8200}"
VAULT_TOKEN="${VAULT_TOKEN}"

# Wait for Vault to be ready
echo "Waiting for Vault to be ready..."
for i in $(seq 1 30); do
  if wget -q --spider --timeout=1 "$VAULT_ADDR/v1/sys/health" 2>/dev/null; then
    echo "Vault is ready"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "Vault did not become ready in time"
    exit 1
  fi
  sleep 1
done

# Export Vault config for CLI
export VAULT_ADDR
export VAULT_TOKEN

echo "Checking if Transit engine is enabled..."
# Check if transit engine is already enabled (idempotent)
if vault secrets list -format=json | grep -q '"transit/"'; then
  echo "Transit engine already enabled"
else
  echo "Enabling Transit engine..."
  vault secrets enable transit
  echo "Transit engine enabled"
fi

# List of Transit keys to create
KEYS="gateway_api service_user_database gateway_recommendation service_popularity_system"

# Create each Transit key if it doesn't exist (idempotent)
for KEY in $KEYS; do
  echo "Checking Transit key: $KEY"

  # Try to read the key; if it exists, this will succeed
  if vault read "transit/keys/$KEY" >/dev/null 2>&1; then
    echo "  ✓ Key '$KEY' already exists"
  else
    echo "  Creating key '$KEY'..."
    vault write -f "transit/keys/$KEY" \
      type=ecdsa-p256 \
      exportable=false \
      allow_plaintext_backup=false
    echo "  ✓ Key '$KEY' created"
  fi
done

echo ""
echo "Vault initialization complete!"
echo "Transit keys available:"
for KEY in $KEYS; do
  echo "  - $KEY"
done

exit 0