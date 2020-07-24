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

	portsToRemove := []int{}

	// if a port was passed as part of the config, it should be the only one
	// removed
	port, configHasPort := getPortFromConfig(config)
	if configHasPort {
		portsToRemove = append(portsToRemove, port)
	}

	existingRegistrations, err := getAllMetricsRegistrations(registrationFetcher, app.Guid)
	for _, registration := range existingRegistrations {
		if config != "" && config != registration.Config {
			continue
		}
		err := removeRegistration(appName, config, registration, cliConn)
		if err != nil {
			return err
		}
		if !configHasPort {
			p, isP := getPortFromConfig(registration.Config)
			if isP {
				portsToRemove = append(portsToRemove, p)
			}
		}
	}

	exposedPorts, err := porter.GetPortsForApp(cliConn, app.Guid)
	if err != nil {
		return err
	}

	// loop over exposedPorts and omit those which need to be removed
	portsForCurl := []int{}
	for _, port := range exposedPorts {
		del := false
		for _, p := range portsToRemove {
			if port == p {
				del = true
				break
			}
		}
		if del == false {
			portsForCurl = append(portsForCurl, port)
		}
	}

	// call setPorts with only ports that need to remain
	err = porter.SetPortsForApp(cliConn, app.Guid, portsForCurl)
	if err != nil {
		return err
	}
	return nil
}

func getPortFromConfig(config string) (int, bool) {
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
