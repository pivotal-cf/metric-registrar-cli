package command

import (
	"strconv"
	"strings"

	porter "github.com/pivotal-cf/metric-registrar-cli/ports"
	"github.com/pivotal-cf/metric-registrar-cli/registrations"
)

func UnregisterLogFormat(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, format string) error {
	return removeRegistrations(registrationFetcher, cliConn, appName, structuredFormat, format)
}

func UnregisterMetricsEndpoint(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, path string) error {
	return removeMetricRegistrations(registrationFetcher, cliConn, appName, metricsEndpoint, path)
}

func removeRegistrations(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, registrationType, config string) error {
	app, err := cliConn.GetApp(appName)
	if err != nil {
		return err
	}
	existingRegistrations, err := registrationFetcher.Fetch(app.Guid, registrationType)

	if err != nil {
		return err
	}

	for _, registration := range existingRegistrations {
		if config != "" && config != registration.Config {
			continue
		}

		err := removeRegistration(appName, config, registration, cliConn)
		if err != nil {
			return err
		}

	}

	return nil
}

func removeMetricRegistrations(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, registrationType, config string) error {
	app, err := cliConn.GetApp(appName)
	if err != nil {
		return err
	}

	existingRegistrations, err := getAllMetricsRegistrations(registrationFetcher, app.Guid)
	if err != nil {
		return err
	}

	portsToRemove, err := removeMatchingRegistrations(existingRegistrations, config, appName, cliConn)

	currentPorts, err := porter.GetPortsForApp(cliConn, app.Guid)
	if err != nil {
		return err
	}

	remainingPorts := getRemainingPorts(currentPorts, portsToRemove)

	// call setPorts with only ports that need to remain
	err = porter.SetPortsForApp(cliConn, app.Guid, remainingPorts)
	if err != nil {
		return err
	}
	return nil
}

func removeMatchingRegistrations(registrations []registrations.Registration, config, appName string, cliConn cliCommandRunner) ([]int, error) {
	portsToRemove := []int{}

	port, configHasPort := getPortFromConfig(config)
	if configHasPort {
		portsToRemove = append(portsToRemove, port)
	}

	for _, registration := range registrations {
		if config != "" && config != registration.Config {
			continue
		}
		err := removeRegistration(appName, config, registration, cliConn)
		if err != nil {
			return nil, err
		}
		if !configHasPort {
			p, isP := getPortFromConfig(registration.Config)
			if isP {
				portsToRemove = append(portsToRemove, p)
			}
		}
	}
	return portsToRemove, nil
}

func getRemainingPorts(currentPorts, portsToRemove []int) []int {
	stay := map[int]bool{}
	for _, p := range currentPorts {
		stay[p] = true
	}
	for _, p := range portsToRemove {
		stay[p] = false
	}

	portsForCurl := []int{}
	for p, ok := range stay {
		if ok {
			portsForCurl = append(portsForCurl, p)
		}
	}

	return portsForCurl
}

func getPortFromConfig(config string) (int, bool) {
	if config == "" {
		return 0, false
	}
	char := string(config[0])
	if char != ":" {
		return 0, false
	}
	a := strings.Split(config, "/")
	b := strings.Trim(a[0], ":")
	i, err := strconv.Atoi(b)
	if err != nil {
		return 0, false
	}
	return i, true
}

func isPort(p string) bool {
	return p != ""
}

func getAllMetricsRegistrations(fetcher registrationFetcher, guid string) ([]registrations.Registration, error) {
	r1, err := fetcher.Fetch(guid, metricsEndpoint)
	if err != nil {
		return nil, err
	}

	r2, err := fetcher.Fetch(guid, secureEndpoint)
	if err != nil {
		return nil, err
	}

	r1 = append(r1, r2...)
	return r1, nil
}

func removeRegistration(appName, config string, registration registrations.Registration, cliConn cliCommandRunner) error {
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
