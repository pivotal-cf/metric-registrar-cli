#!/usr/bin/env bash
set -exu

pushd "$(dirname $0)/.."
  go get -t ./...
  ${GOPATH}/bin/ginkgo -r --race --randomizeAllSpecs
popd