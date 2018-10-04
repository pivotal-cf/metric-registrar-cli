package main

import (
    "code.cloudfoundry.org/cli/plugin"
    "github.com/pivotal-cf/metric-registrar-cli/command"
)

func main() {
    plugin.Start(command.MetricRegistrarCli{})
}
