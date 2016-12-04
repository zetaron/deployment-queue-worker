#!/bin/sh
set -e

export DEPLOYMENT_CACHE_VOLUME_NAME=$1
export SECRET_VOLUME_NAME=$2
export WORKER_IMAGE=$3

stdin=$(cat)

docker volume create --name $DEPLOYMENT_CACHE_VOLUME_NAME
echo "$stdin" | docker run --rm -i -v $DEPLOYMENT_CACHE_VOLUME_NAME:/var/cache/deployment alpine:3.4 dd of=/var/cache/deployment/deployment-event.json
docker run --rm -v $DEPLOYMENT_CACHE_VOLUME_NAME:/var/cache/deployment -v $SECRET_VOLUME_NAME:/var/cache/secrets -e DEPLOYMENT_CACHE_VOLUME_NAME -e SECRET_VOLUME_NAME $WORKER_IMAGE
