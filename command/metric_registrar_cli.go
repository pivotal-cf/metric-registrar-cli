package command

import (
    "code.cloudfoundry.org/cli/plugin"
    "code.cloudfoundry.org/cli/plugin/models"
    "github.com/jessevdk/go-flags"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"

    "fmt"
    "os"
)

const (
    pluginName = "metric-registrar"

    structuredFormat = "structured-format"
    metricsEndpoint  = "metrics-endpoint"
)

type MetricRegistrarCli struct {
    Major int
    Minor int
    Patch int
}

type registrationFetcher interface {
    Fetch(string, string) ([]registrations.Registration, error)
    FetchAll(string) (map[string][]registrations.Registration, error)
}

type cliCommandRunner interface {
    CliCommandWithoutTerminalOutput(...string) ([]string, error)
    GetServices() ([]plugin_models.GetServices_Model, error)
    GetApp(string) (plugin_models.GetAppModel, error)
    GetApps() ([]plugin_models.GetAppsModel, error)
}

func (c MetricRegistrarCli) Run(cliConnection plugin.CliConnection, args []string) {
    command := command(args)
    parseArgs(command, args)

    exitIfErr(command.Run(registrations.NewFetcher(cliConnection), cliConnection))
}

func command(args []string) Command {
    commandName := args[0]
    if commandName == "CLI-MESSAGE-UNINSTALL" {
        os.Exit(0)
    }

    command, ok := Registry[commandName]
    if !ok {
        fmt.Println("unknown command")
        os.Exit(1)
    }

    return command
}

func parseArgs(command Command, args []string) {
    parser := flags.NewParser(command.Flags, flags.HelpFlag)

    remainingArgs, err := parser.ParseArgs(args[1:])
    if err != nil {
        exitUsage(err.Error(), command.Usage())
    }

    if len(remainingArgs) != 0 {
        exitUsage("too many arguments", command.Usage())
    }
}

func exitUsage(message, usage string) {
    fmt.Printf("incorrect usage: %s\n\n", message)
    fmt.Println(usage)
    os.Exit(1)
}

func exitIfErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
