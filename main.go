package main

import (
    "code.cloudfoundry.org/cli/plugin"
    "github.com/pivotal-cf/prism-cli/command"
)

func main() {
    plugin.Start(command.PrismCli{})
}
