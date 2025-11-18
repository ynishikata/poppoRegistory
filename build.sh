#!/bin/bash
set -e

# Build frontend
echo "Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Build backend
echo "Building backend..."
go mod download
go build -o bin/server .

echo "Build complete!"

