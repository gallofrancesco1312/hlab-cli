#!/usr/bin/env bash

set -e

mkdir -p certs
cd certs

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

echo "Generating server key..."

openssl genrsa -out server.key 2048

cat > server.ext <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth

subjectAltName=@alt_names

[alt_names]
DNS.1=localhost
IP.1=127.0.0.1
IP.2=::1
EOF

echo "Generating CSR..."

openssl req \
  -new \
  -key server.key \
  -out server.csr \
  -subj "/C=IT/ST=Lombardy/L=Milan/O=Local Dev/CN=localhost"

echo "Signing certificate..."

openssl x509 \
  -req \
  -in server.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out server.crt \
  -days 825 \
  -sha256 \
  -extfile server.ext

echo
echo "Verifying..."

openssl verify -CAfile ca.crt server.crt

echo
echo "Done."
echo
echo "Files created:"
echo "  certs/ca.crt"
echo "  certs/ca.key"
echo "  certs/server.crt"
echo "  certs/server.key"