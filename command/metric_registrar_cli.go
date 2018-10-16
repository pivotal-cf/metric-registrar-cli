package command

import (
    "code.cloudfoundry.org/cli/plugin"
    "code.cloudfoundry.org/cli/plugin/models"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"

    "fmt"
    "os"
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

func exitIfErr(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
