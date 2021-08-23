#!/usr/bin/env bash

set -e
cf uninstall-plugin metric-registrar || true # suppress errors

cd "$(dirname $0)/.."
CLI_BUILD_OS=linux CLI_BUILD_ARCH=amd64 ./scripts/build.sh

cf install-plugin ./plugins/metric-registrar-cli-linux-amd64-* -f
