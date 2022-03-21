#!/usr/bin/env bash
set -exu

pushd "$(dirname $0)/.."
  go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --keep-going --race --trace
popd
