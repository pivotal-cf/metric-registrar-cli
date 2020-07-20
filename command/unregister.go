package command

import (
	"github.com/pivotal-cf/metric-registrar-cli/registrations"
)

func UnregisterLogFormat(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, format string) error {
	return removeRegistrations(registrationFetcher, cliConn, appName, structuredFormat, format)
}

func UnregisterMetricsEndpoint(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, path string) error {
	return removeRegistrations(registrationFetcher, cliConn, appName, metricsEndpoint, path)
}

func removeRegistrations(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, registrationType, config string) error {
	existingRegistrations, err := existingRegistrations(registrationFetcher, cliConn, appName, registrationType)
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

func existingRegistrations(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName string, registrationType string) ([]registrations.Registration, error) {
	app, err := cliConn.GetApp(appName)
	if err != nil {
		return nil, err
	}

	return registrationFetcher.Fetch(app.Guid, registrationType)
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
