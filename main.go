package main

import (
    "code.cloudfoundry.org/cli/plugin"
    "github.com/pivotal-cf/cf-metric-registrar/command"
)

func main() {
    plugin.Start(command.MetricRegistrarCli{})
}
