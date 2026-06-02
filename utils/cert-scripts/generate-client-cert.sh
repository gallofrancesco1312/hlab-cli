#!/usr/bin/env bash

set -e

CERT_DIR="certs"

if [ ! -f "$CERT_DIR/ca.crt" ] || [ ! -f "$CERT_DIR/ca.key" ]; then
    echo "ERROR: $CERT_DIR/ca.crt o $CERT_DIR/ca.key non trovati"
    exit 1
fi

cd "$CERT_DIR"

echo "Generating client key..."

openssl genrsa -out client.key 2048

echo "Generating client CSR..."

openssl req \
  -new \
  -key client.key \
  -out client.csr \
  -subj "/C=IT/ST=Lombardy/L=Milan/O=Local Dev/CN=test-client"

cat > client.ext <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
EOF

echo "Signing client certificate..."

openssl x509 \
  -req \
  -in client.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out client.crt \
  -days 825 \
  -sha256 \
  -extfile client.ext

echo "Verifying..."

openssl verify -CAfile ca.crt client.crt

rm -f client.csr client.ext

echo
echo "Done."
echo
echo "Generated files:"
echo "  certs/client.crt"
echo "  certs/client.key"