#!/bin/bash
set -e

echo "🔨 Building CardForge..."
echo ""

# Build frontend
echo "📦 Building frontend..."
cd web
npm run build
cd ..
echo "✅ Frontend built successfully"
echo ""

# Build Go binary
echo "🚀 Building Go binary..."
go build -ldflags="-s -w" -o cardforge cmd/server/main.go
echo "✅ Binary built successfully"
echo ""

echo "🎉 Build complete!"
echo "Binary location: ./cardforge"
echo ""
echo "To run:"
echo "  ./cardforge"
