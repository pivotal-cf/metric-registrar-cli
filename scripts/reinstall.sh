#!/usr/bin/env bash

set -e
cf uninstall-plugin pm-please-add-details || true # suppress errors

cd "$(dirname $0)/../src/cliplugin"
go build -o bin/cli

cf install-plugin ./bin/cli -f
