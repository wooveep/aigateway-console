#!/bin/sh
set -eu

IMAGE_NAME="${IMAGE_NAME:-aigateway-console:0.0.1}"
BUILD_ARGS=""
MAVEN_ARGS="${MAVEN_ARGS:-}"
SKIP_TESTS="${SKIP_TESTS:-true}"
SKIP_STATIC_CHECKS="${SKIP_STATIC_CHECKS:-true}"

if [ -n "${VERSION:-}" ]; then
    BUILD_ARGS="$BUILD_ARGS -Dapp.build.version=$VERSION"
fi

if [ -n "${DEV:-}" ]; then
    BUILD_ARGS="$BUILD_ARGS -Dapp.build.dev=$DEV"
fi

if [ "$SKIP_TESTS" = "true" ]; then
    MAVEN_ARGS="$MAVEN_ARGS -Dmaven.test.skip=true"
fi

if [ "$SKIP_STATIC_CHECKS" = "true" ]; then
    MAVEN_ARGS="$MAVEN_ARGS -Dpmd.skip=true -Dcheckstyle.skip=true"
fi

./mvnw clean package -Dpmd.language=en $MAVEN_ARGS $BUILD_ARGS
docker build -t "$IMAGE_NAME" -f Dockerfile .
