#!/usr/bin/env bash

set -e

TYPE="${1:-}"
CERT_DIR="${2:-./certs}"

if [[ "$TYPE" != "server" && "$TYPE" != "client" ]]; then
  echo "Usage: $0 <server|client> [cert-dir]"
  exit 1
fi

mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

if [[ -f ca.crt && -f ca.key ]]; then
  echo "Reusing existing CA."
else
  echo "Generating CA..."
  openssl genrsa -out ca.key 4096
  openssl req \
    -x509 \
    -new \
    -nodes \
    -key ca.key \
    -sha256 \
    -days 3650 \
    -out ca.crt \
    -subj "/C=IT/ST=Lombardy/L=Milan/O=Local Dev/CN=Local Development CA"
fi

echo "Generating $TYPE key..."

openssl genrsa -out "$TYPE.key" 2048

if [[ "$TYPE" == "server" ]]; then
  cat > "$TYPE.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth

subjectAltName=@alt_names

[alt_names]
DNS.1=localhost
IP.1=192.168.1.19
IP.2=::1
EOF
  CN="localhost"
else
  cat > "$TYPE.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
EOF
  CN="test-client"
fi

echo "Generating CSR..."

openssl req \
  -new \
  -key "$TYPE.key" \
  -out "$TYPE.csr" \
  -subj "/C=IT/ST=Lombardy/L=Milan/O=Local Dev/CN=$CN"

echo "Signing certificate..."

openssl x509 \
  -req \
  -in "$TYPE.csr" \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out "$TYPE.crt" \
  -days 825 \
  -sha256 \
  -extfile "$TYPE.ext"

rm -f "$TYPE.csr" "$TYPE.ext"

echo
echo "Verifying..."
openssl verify -CAfile ca.crt "$TYPE.crt"

echo
echo "Done."
echo
echo "Files created in $CERT_DIR:"
echo "  ca.crt  ca.key"
echo "  $TYPE.crt  $TYPE.key"
