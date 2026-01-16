#!/bin/sh

set -euo pipefail

CERT_DIR="${CERT_DIR:-/certs}"
DAYS="${DAYS:-3650}"
CN="${CN:-localhost}"

mkdir -p "$CERT_DIR"

# 1) CA key + cert
openssl genrsa -out "$CERT_DIR/ca.key" 4096
openssl req -x509 -new -nodes -key "$CERT_DIR/ca.key" \
  -sha256 -days "$DAYS" -out "$CERT_DIR/ca.crt" \
  -subj "/CN=pow-ddos-guard-ca"

# 2) Server key + CSR
openssl genrsa -out "$CERT_DIR/server.key" 2048
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
  -subj "/CN=$CN"

# 3) SAN 
cat > "$CERT_DIR/san.ext" <<EOF
subjectAltName=DNS:localhost,DNS:server,IP:127.0.0.1
extendedKeyUsage=serverAuth
keyUsage=digitalSignature,keyEncipherment
EOF

# 4) Sign server cert with CA
openssl x509 -req -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial \
  -out "$CERT_DIR/server.crt" -days "$DAYS" -sha256 \
  -extfile "$CERT_DIR/san.ext"

# cleanup
rm -f "$CERT_DIR/server.csr" "$CERT_DIR/san.ext"

# permissions 
chown 10001:10001 /certs/*.key /certs/*.crt
chmod 600 "$CERT_DIR"/*.key
chmod 644 "$CERT_DIR"/*.crt
