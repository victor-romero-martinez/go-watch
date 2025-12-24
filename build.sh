#!/bin/bash

# --- CONFIGURACIÓN ---
mkdir -p builds
PACKAGE_NAME="gow"
VERSION="0.0.2"

echo "Building MMP binaries (Version: $VERSION) for multiple platforms..."

# Parámetros de flags para la inyección de versión
LDFLAG_VALUE="-X main.version=${VERSION}"

# --- BUILDS ---

# 1. Linux (x86-64)
echo "-> Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o dist/$PACKAGE_NAME-linux-amd64 -ldflags "$LDFLAG_VALUE" main.go

# 2. Windows (x86-64)
echo "-> Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o dist/$PACKAGE_NAME-windows-amd64.exe -ldflags "$LDFLAG_VALUE" main.go

# 3. Apple Silicon (M1, M2, M3, etc.)
echo "-> macOS (arm64/Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o dist/$PACKAGE_NAME-mac-arm64 -ldflags "$LDFLAG_VALUE" main.go

# 4. Intel/AMD 64-bit
echo "-> macOS (amd64/Intel)..."
GOOS=darwin GOARCH=amd64 go build -o dist/$PACKAGE_NAME-mac-intel -ldflags "$LDFLAG_VALUE" main.go

# 5. Raspberry Pi / ARM (64-bit)
# echo "-> Linux (arm64/Raspberry Pi)..."
# GOOS=linux GOARCH=arm64 go build -o dist/$PACKAGE_NAME-linux-arm64 -ldflags "$LDFLAG_VALUE" main.go

echo "✅ Builds complete in the 'dist/' directory."