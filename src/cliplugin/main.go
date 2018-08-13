package main

import (
    "code.cloudfoundry.org/cli/plugin"
    "cliplugin/command"
)

func main() {
    plugin.Start(command.PrismCli{})
}
