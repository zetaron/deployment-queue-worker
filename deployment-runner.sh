#!/bin/sh
set -e

stdin=$(</dev/stdin)
docker volume create $DEPLOYMENT_CACHE_VOLUME_NAME
echo "$stdin" | docker run --rm -i -v $DEPLOYMENT_CACHE_VOLUME_NAME:/var/cache/deployment alpine:3.4 dd of=/var/cache/deployment/deployment-event.json
docker-compose --file docker-compose.deploy.yml --project-name $DEPLOYMENT_CACHE_VOLUME_NAME up -d
docker wait "${DEPLOYMENT_CACHE_VOLUME_NAME}_deploy"
