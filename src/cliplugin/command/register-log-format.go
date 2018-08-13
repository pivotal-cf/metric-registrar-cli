package command

import (
    "code.cloudfoundry.org/cli/plugin"
    "code.cloudfoundry.org/cli/plugin/models"
    "fmt"
    "os"
)

const pluginName = "pm-please-add-details" // also in scripts/reinstall.sh
const registerLogFormatCommand = "register-log-format"
const registerLogFormatUsage = "cf register-log-format APPNAME FORMAT"

type PrismCli struct {}

type cliCommandRunner interface {
    CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
    GetServices() ([]plugin_models.GetServices_Model, error)
}

func (c PrismCli) Run(cliConnection plugin.CliConnection, args []string) {
    switch args[0] {
    case registerLogFormatCommand:
        RegisterLogFormat(cliConnection, args[1:])
    case "CLI-MESSAGE-UNINSTALL":
        // do nothing
    }
}

func (c PrismCli) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name: pluginName,
        Version: plugin.VersionType{
            Major: 0,
            Minor: 0,
            Build: 0,
        },
        Commands: []plugin.Command{
            {
                Name:     registerLogFormatCommand,
                HelpText: "This will register bound applications so that structured logs of the given format can be parsed",
                UsageDetails: plugin.Usage{
                    Usage: registerLogFormatUsage,
                },
            },
        },
    }
}

func RegisterLogFormat(cliConn cliCommandRunner, args []string) {
    if len(args) != 2 {
        panic("Usage: " + registerLogFormatUsage)
    }
    appName := args[0]
    logFormat := args[1]
    serviceName := "structured-format-" + logFormat
    exists, err := findExistingService(cliConn, serviceName)
    exitIfErr(err)

    if !exists {
        binding := "structured-format://" + logFormat
        _, err := cliConn.CliCommandWithoutTerminalOutput("create-user-provided-service", serviceName, "-l", binding)
        exitIfErr(err)
    }

    _, err = cliConn.CliCommandWithoutTerminalOutput("bind-service", appName, serviceName)
    exitIfErr(err)
}

func findExistingService(cliConn cliCommandRunner, serviceName string) (bool, error) {
    existingServices, err := cliConn.GetServices()
    if err != nil {
        return false, err
    }
    for _, s := range existingServices {
        if s.Name == serviceName {
            return true, nil
        }
    }
    return false, nil
}

func exitIfErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
