#!/bin/bash

# Lambra Docker Image Build Script
# Usage: ./scripts/build-images.sh [version]

set -e

VERSION=${1:-latest}
REGISTRY=${DOCKER_REGISTRY:-"lambra"}

echo "Building Lambra Docker images (version: $VERSION)"
echo "================================================"

# Build Backend
echo ""
echo "[1/2] Building backend image..."
docker build \
  -t ${REGISTRY}/backend:${VERSION} \
  -t ${REGISTRY}/backend:latest \
  -f backend/Dockerfile \
  backend/

echo "✓ Backend image built: ${REGISTRY}/backend:${VERSION}"

# Build Frontend
echo ""
echo "[2/2] Building frontend image..."
docker build \
  -t ${REGISTRY}/frontend:${VERSION} \
  -t ${REGISTRY}/frontend:latest \
  -f frontend/Dockerfile \
  frontend/

echo "✓ Frontend image built: ${REGISTRY}/frontend:${VERSION}"

echo ""
echo "================================================"
echo "Build complete!"
echo ""
echo "Images created:"
echo "  - ${REGISTRY}/backend:${VERSION}"
echo "  - ${REGISTRY}/frontend:${VERSION}"
echo ""
echo "To push to registry:"
echo "  docker push ${REGISTRY}/backend:${VERSION}"
echo "  docker push ${REGISTRY}/frontend:${VERSION}"
echo ""
echo "To test locally:"
echo "  cd dist && docker-compose up -d"
