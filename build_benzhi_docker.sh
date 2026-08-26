#!/bin/bash
set -e
IMAGE_NAME=${1:-my-project}
DOCKER_PLATFORM=${2:-linux/amd64}
docker build --platform $DOCKER_PLATFORM -f benzhi.Dockerfile -t $IMAGE_NAME .
echo "Docker image '$IMAGE_NAME' built successfully."
echo "  docker run -d -P --name fresnel-tm-b14 $IMAGE_NAME:latest"
echo "  curl -s http://127.0.0.1:\$(docker port fresnel-tm-b14 8080 | cut -d: -f2)/api/examples"
echo "  docker rm -f fresnel-tm-b14"
