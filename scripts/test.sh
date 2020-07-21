#!/usr/bin/env bash
set -exu

pushd "$(dirname $0)/.."
  go mod download
  ${GOPATH}/bin/ginkgo -r --race --randomizeAllSpecs
popd
