package command

import (
    "errors"
    "github.com/jessevdk/go-flags"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"
)

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
