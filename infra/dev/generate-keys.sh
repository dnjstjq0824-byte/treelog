#!/bin/sh
# Generates local-only development secrets for TreeBlog.
# Writes ./infra/dev/jwt_rsa (RS256 private key) and prints .env entries.
set -eu

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DEV_DIR="$ROOT/infra/dev"
mkdir -p "$DEV_DIR"

KEY_PATH="$DEV_DIR/jwt_rsa"
if [ ! -f "$KEY_PATH" ]; then
  openssl genrsa -out "$KEY_PATH" 2048
  chmod 600 "$KEY_PATH"
  echo "wrote $KEY_PATH"
else
  echo "kept existing $KEY_PATH"
fi

PEPPER="$(openssl rand -base64 32)"
cat <<EOF

Add the following to your .env (values are LOCAL ONLY, never commit .env):

TREEBLOG_JWT_PRIVATE_KEY_PATH=/run/secrets/treeblog/jwt_rsa
TREEBLOG_CODE_PEPPER=$PEPPER

For host-side runs (go run ./cmd/api), use instead:
TREEBLOG_JWT_PRIVATE_KEY_PATH=$KEY_PATH
EOF
