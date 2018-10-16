package command

import (
    "code.cloudfoundry.org/cli/plugin"
    "code.cloudfoundry.org/cli/plugin/models"
    "github.com/jessevdk/go-flags"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"

    "errors"
    "fmt"
    "os"
    "strings"
)

const (
    pluginName                       = "metric-registrar"
    registerLogFormatCommand         = "register-log-format"
    registerMetricsEndpointCommand   = "register-metrics-endpoint"
    unregisterLogFormatCommand       = "unregister-log-format"
    unregisterMetricsEndpointCommand = "unregister-metrics-endpoint"

    registerLogFormatUsage         = "cf register-log-format APPNAME <json|DogStatsD>"
    registerMetricsEndpointUsage   = "cf register-metrics-endpoint APPNAME PATH"
    unregisterLogFormatUsage       = "cf unregister-log-format APPNAME [-f FORMAT]"
    unregisterMetricsEndpointUsage = "cf unregister-metrics-endpoint APPNAME [-p PATH]"

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
    case unregisterLogFormatCommand:
        registrationFetcher := registrations.NewFetcher(cliConnection)
        err := UnregisterLogFormat(registrationFetcher, cliConnection, args[1:])
        exitIfErr(err)
    case unregisterMetricsEndpointCommand:
        registrationFetcher := registrations.NewFetcher(cliConnection)
        err := UnregisterMetricsEndpoint(registrationFetcher, cliConnection, args[1:])
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
            {
                Name:     unregisterLogFormatCommand,
                HelpText: "Unregister log formats",
                UsageDetails: plugin.Usage{
                    Usage: unregisterLogFormatUsage,
                    Options: map[string]string{
                        "-f": "unregister only the specified log format",
                    },
                },
            },
            {
                Name:     unregisterMetricsEndpointCommand,
                HelpText: "Unregister metrics endpoints",
                UsageDetails: plugin.Usage{
                    Usage: unregisterMetricsEndpointUsage,
                    Options: map[string]string{
                        "-p": "unregister only the specified path",
                    },
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

    return EnsureServiceAndBind(cliConn, appName, structuredFormat, logFormat)
}

func RegisterMetricsEndpoint(cliConn cliCommandRunner, args []string) error {
    if len(args) != 2 {
        return errors.New("usage: " + registerMetricsEndpointUsage)
    }
    appName := args[0]
    path := args[1]

    return EnsureServiceAndBind(cliConn, appName, metricsEndpoint, path)
}

func UnregisterLogFormat(registrationFetcher registrationFetcher, cliConn cliCommandRunner, args []string) error {
    type opts struct {
        Format string `short:"f" long:"format"`
    }

    var options opts
    argsNoFlags, err := flags.ParseArgs(&options, args)
    if err != nil {
        return err
    }

    if len(argsNoFlags) != 1 {
        return errors.New("usage: " + unregisterLogFormatUsage)
    }
    appName := argsNoFlags[0]

    return removeRegistrations(appName, structuredFormat, options.Format, cliConn, registrationFetcher)

}

func UnregisterMetricsEndpoint(registrationFetcher registrationFetcher, cliConn cliCommandRunner, args []string) error {
    type opts struct {
        Path string `short:"p" long:"path"`
    }

    var options opts
    argsNoFlags, err := flags.ParseArgs(&options, args)
    if err != nil {
        return err
    }

    if len(argsNoFlags) != 1 {
        return errors.New("usage: " + unregisterMetricsEndpointUsage)
    }
    appName := argsNoFlags[0]

    return removeRegistrations(appName, metricsEndpoint, options.Path, cliConn, registrationFetcher)
}

func removeRegistrations(appName, registrationType, config string, cliConn cliCommandRunner, registrationFetcher registrationFetcher) error {
    app, err := cliConn.GetApp(appName)
    if err != nil {
        return err
    }

    existingRegistrations, err := registrationFetcher.Fetch(app.Guid, registrationType)
    if err != nil {
        return err
    }

    for _, registration := range existingRegistrations {
        err := removeRegistration(appName, config, registration, cliConn)
        if err != nil {
            return err
        }
    }

    return nil
}

func removeRegistration(appName, config string, registration registrations.Registration, cliConn cliCommandRunner) error {
    if config != "" && config != registration.Config {
        return nil
    }

    _, err := cliConn.CliCommandWithoutTerminalOutput("unbind-service", appName, registration.Name)
    if err != nil {
        return err
    }

    if registration.NumberOfBindings == 1 {
        _, err = cliConn.CliCommandWithoutTerminalOutput("delete-service", registration.Name, "-f")
        if err != nil {
            return err
        }
    }

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
