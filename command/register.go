package command

import (
    "strings"
)

func RegisterLogFormat(cliConn cliCommandRunner,  appName, logFormat string) error {
    return ensureServiceAndBind(cliConn, appName, structuredFormat, logFormat)
}

func RegisterMetricsEndpoint(cliConn cliCommandRunner, appName, path string) error {
    return ensureServiceAndBind(cliConn, appName, metricsEndpoint, path)
}

func ensureServiceAndBind(cliConn cliCommandRunner, appName, serviceProtocol, config string) error {
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