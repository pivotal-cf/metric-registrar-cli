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

func UnregisterMetricsEndpoint(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, path, port string) error {
	return removeMetricRegistrations(registrationFetcher, cliConn, appName, metricsEndpoint, path, port)
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

		err := removeRegistration(appName, registration, cliConn)
		if err != nil {
			return err
		}

	}

	return nil
}

func removeMetricRegistrations(registrationFetcher registrationFetcher, cliConn cliCommandRunner, appName, registrationType, path, port string) error {
	config := path
	if port != "" {
		config = ":" + port + config
	}
	app, err := cliConn.GetApp(appName)
	if err != nil {
		return err
	}

	existingRegistrations, err := getAllMetricsRegistrations(registrationFetcher, app.Guid)
	if err != nil {
		return err
	}

	portsToRemove, err := removeMatchingRegistrations(existingRegistrations, config, appName, cliConn)
	if err != nil {
		return err
	}

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

func configMatch(config1, config2 string) bool {
	if config1 == "" {
		return true
	}

	return config1 == config2
}

func removeMatchingRegistrations(registrations []registrations.Registration, config, appName string, cliConn cliCommandRunner) ([]int, error) {
	keepPort := map[int]bool{}

	for _, r := range registrations {
		p := getPortFromConfig(r.Config)

		if !configMatch(config, r.Config) {
			keepPort[p] = true
			continue
		}

		err := removeRegistration(appName, r, cliConn)
		if err != nil {
			return nil, err
		}

		if !keepPort[p] {
			keepPort[p] = false
		}
	}

	remove := []int{}
	for p, keep := range keepPort {
		if !keep {
			remove = append(remove, p)
		}
	}

	return remove, nil
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

func getPortFromConfig(config string) int {
	if config == "" {
		return -1
	}
	char := string(config[0])
	if char != ":" {
		return -1
	}
	a := strings.Split(config, "/")
	b := strings.Trim(a[0], ":")
	i, err := strconv.Atoi(b)
	if err != nil {
		return -1
	}
	return i
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

func removeRegistration(appName string, registration registrations.Registration, cliConn cliCommandRunner) error {
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
