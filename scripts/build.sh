#!/usr/bin/env bash
set -e

pushd() {
  command pushd $@ >/dev/null
}

popd() {
  command popd $@ >/dev/null
}

pushd $(dirname $(dirname $0))
  PROJECT_ROOT=$(pwd -P)
popd

PLUGIN_OUTPUT=${PLUGIN_DIRECTORY:-${PROJECT_ROOT}/plugins}
mkdir -p ${PLUGIN_OUTPUT}
pushd ${PLUGIN_OUTPUT}
  PLUGIN_OUTPUT=$(pwd -P)
popd

function build_for() {
  local OS=$1
  local ARCH=$2

  echo "Building CLI for $OS $ARCH..."

  BUILD_NUMBER=${VERSION:-0.0.1}
  MAJOR_MINOR_PATCH=( ${BUILD_NUMBER//./ })
  VERSION_FLAGS="-X main.Major=${MAJOR_MINOR_PATCH[0]} -X main.Minor=${MAJOR_MINOR_PATCH[1]} -X main.Patch=${MAJOR_MINOR_PATCH[2]}"

  GOARCH=${ARCH} GOOS=${OS} go build -o ${PLUGIN_OUTPUT}/metric-registrar-cli-${OS}-${ARCH}-${BUILD_NUMBER} -ldflags "${VERSION_FLAGS}"
}

pushd ${PROJECT_ROOT}
  if [[ -z "${NO_DEP:-}" ]]; then
    echo "Fetching dependencies..."
    go get ./...
  fi

  if [[ -n "${CLI_BUILD_OS}" ]]; then
    build_for ${CLI_BUILD_OS} ${CLI_BUILD_ARCH:-amd64}
  else
    for OS in windows linux darwin; do
      build_for ${OS} amd64
    done
    for OS in windows linux; do
      build_for ${OS} 386
    done
  fi
popd