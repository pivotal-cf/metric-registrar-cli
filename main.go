package main

import (
	"strconv"

	"github.com/pivotal-cf/metric-registrar-cli/command"

	"code.cloudfoundry.org/cli/plugin"
)

var Major string
var Minor string
var Patch string

func main() {
	plugin.Start(command.MetricRegistrarCli{
		Major: getIntOrPanic(Major),
		Minor: getIntOrPanic(Minor),
		Patch: getIntOrPanic(Patch),
	})
}

func getIntOrPanic(toInt string) int {
	theInt, err := strconv.Atoi(toInt)
	if err != nil {
		panic("unable to parse version: " + err.Error())
	}
	return theInt
}
