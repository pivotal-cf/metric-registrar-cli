#!/usr/bin/env bash

set -e
cf uninstall-plugin app-metric-registrar || true # suppress errors

cd "$(dirname $0)/.."
go build -o bin/cli

cf install-plugin ./bin/cli -f
