#!/bin/bash

echo "Building Zendo Todo App..."

# Build frontend
echo "Building frontend..."
cd zendo-frontend
npm run build
cd ..

# Copy frontend build to backend static directory
echo "Copying frontend build to backend..."
rm -rf zendo-backend/static
cp -r zendo-frontend/build zendo-backend/static

# Build backend
echo "Building backend..."
cd zendo-backend
go mod tidy
go build -o zendo .

echo "Build complete! Run ./zendo-backend/zendo to start the application." 