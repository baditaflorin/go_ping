#!/bin/bash
set -e

# Default image name
IMAGE="${IMAGE:-ghcr.io/baditaflorin/ping:latest}"

echo "Building multi-platform Docker image: $IMAGE"
echo "Platforms: linux/amd64, linux/arm64/v8"
echo

# Build and push multi-platform image
docker buildx build \
  --platform linux/amd64,linux/arm64/v8 \
  -t "$IMAGE" \
  --push \
  .

echo
echo "✓ Successfully built and pushed: $IMAGE"
