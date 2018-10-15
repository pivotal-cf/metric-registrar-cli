package command

import (
    "code.cloudfoundry.org/cli/plugin"
    "code.cloudfoundry.org/cli/plugin/models"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"

    "errors"
    "fmt"
    "os"
    "strings"
)

const pluginName = "metric-registrar"
const registerLogFormatCommand = "register-log-format"
const registerMetricsEndpointCommand = "register-metrics-endpoint"
const registerLogFormatUsage = "cf register-log-format APPNAME <json|DogStatsD>"
const registerMetricsEndpointUsage = "cf register-metrics-endpoint APPNAME PATH"

type MetricRegistrarCli struct {
    Major int
    Minor int
    Patch int
}

type registrationFetcher interface {
    Fetch(string, string) ([]registrations.Registration, error)
}

type cliCommandRunner interface {
    CliCommandWithoutTerminalOutput(...string) ([]string, error)
    GetServices() ([]plugin_models.GetServices_Model, error)
    GetApp(string) (plugin_models.GetAppModel, error)
}

func (c MetricRegistrarCli) Run(cliConnection plugin.CliConnection, args []string) {
    switch args[0] {
    case registerLogFormatCommand:
        err := RegisterLogFormat(cliConnection, args[1:])
        exitIfErr(err)
    case registerMetricsEndpointCommand:
        err := RegisterMetricsEndpoint(cliConnection, args[1:])
        exitIfErr(err)
    case "CLI-MESSAGE-UNINSTALL":
        // do nothing
    }
}

func (c MetricRegistrarCli) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name: pluginName,
        Version: plugin.VersionType{
            Major: c.Major,
            Minor: c.Minor,
            Build: c.Patch,
        },
        Commands: []plugin.Command{
            {
                Name:     registerLogFormatCommand,
                HelpText: "Register bound applications so that structured logs of the given format can be parsed",
                UsageDetails: plugin.Usage{
                    Usage: registerLogFormatUsage,
                },
            },
            {
                Name:     registerMetricsEndpointCommand,
                HelpText: "Register a metrics endpoint which will be scraped at the interval defined at deploy",
                UsageDetails: plugin.Usage{
                    Usage: registerMetricsEndpointUsage,
                },
            },
        },
    }
}

func RegisterLogFormat(cliConn cliCommandRunner, args []string) error {
    if len(args) != 2 {
        return errors.New("usage: " + registerLogFormatUsage)
    }
    appName := args[0]
    logFormat := args[1]

    return EnsureServiceAndBind(cliConn, appName, "structured-format", logFormat)
}

func RegisterMetricsEndpoint(cliConn cliCommandRunner, args []string) error {
    if len(args) != 2 {
        return errors.New("usage: " + registerMetricsEndpointUsage)
    }
    appName := args[0]
    path := args[1]

    return EnsureServiceAndBind(cliConn, appName, "metrics-endpoint", path)
}

func UnregisterLogFormat(registrationFetcher registrationFetcher, cliConn cliCommandRunner, args []string) error {
    //existingServices, err := registrationFetcher.GetServices()
    //if err != nil {
    //    //TODO
    //}
    //appName := args[0]
    //
    //for _, s := range existingServices {
    //    cliConn.CliCommandWithoutTerminalOutput("unbind-service", appName, s.Name) //TODO err
    //    if len(s.ApplicationNames) == 1 && s.ApplicationNames[0] == appName {
    //        cliConn.CliCommandWithoutTerminalOutput("delete-service", s.Name, "-f") //TODO err
    //    }
    //}
    return nil
}

//TODO shouldn't be exported
func EnsureServiceAndBind(cliConn cliCommandRunner, appName, serviceProtocol, config string) error {
    cleanedConfig := strings.Trim(strings.Replace(config, "/", "-", -1), "-")
    serviceName := serviceProtocol + "-" + cleanedConfig
    exists, err := findExistingService(cliConn, serviceName)
    if err != nil {
        return err
    }

    if !exists {
        binding := serviceProtocol + "://" + config
        _, err := cliConn.CliCommandWithoutTerminalOutput("create-user-provided-service", serviceName, "-l", binding)
        if err != nil {
            return err
        }
    }

    _, err = cliConn.CliCommandWithoutTerminalOutput("bind-service", appName, serviceName)

    return err
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
