package command

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	pluginmodels "code.cloudfoundry.org/cli/plugin/models"
)

func RegisterLogFormat(cliConn cliCommandRunner, appName, logFormat string) error {
	return ensureServiceAndBind(cliConn, appName, structuredFormat, logFormat)
}

func RegisterMetricsEndpoint(cliConn cliCommandRunner, appName, route, internalPort string) error {
	app, err := cliConn.GetApp(appName)
	if err != nil {
		return err
	}

	// if they provide a full URL rather than a relative path, we need to
	// verify that the route is actually associated with the app
	// (we don't want people trying to scrape https://cia.gov/metrics)
	if route[0] != '/' {
		err = validateRouteForApp(route, app)
		if err != nil {
			return err
		}
	}

	serviceProtocol := metricsEndpoint
	if internalPort != "" {
		route = ":" + internalPort + route
		serviceProtocol = secureEndpoint
		port, err := strconv.Atoi(internalPort)
		if err != nil {
			return err
		}
		err = exposePortForApp(cliConn, app.Guid, port)
		if err != nil {
			return err
		}
	}

	return ensureServiceAndBind(cliConn, appName, serviceProtocol, route)
}

func validateRouteForApp(requestedRoute string, app pluginmodels.GetAppModel) error {
	requested, err := url.Parse(ensureHttpsPrefix(requestedRoute))
	if err != nil {
		return fmt.Errorf("unable to parse requested route: %s", err)
	}
	for _, r := range app.Routes {
		var host string
		host = formatHost(r)
		route := &url.URL{
			Host: host,
			Path: "/" + r.Path,
		}

		if requested.Host == route.Host && strings.HasPrefix(requested.Path, route.Path) {
			return nil
		}
	}
	return fmt.Errorf("route '%s' is not bound to app '%s'", requestedRoute, app.Name)
}

func exposePortForApp(cliConn cliCommandRunner, guid string, port int) error {
	existingPorts, err := getPortsForApp(cliConn, guid)
	if err != nil {
		return err
	}

	// don't need to make a PUT request if it's already exposed
	for _, p := range existingPorts {
		if p == port {
			return nil
		}
	}

	newPorts := append(existingPorts, port)
	return setPortsForApp(cliConn, guid, newPorts)
}

/*********************** TODO: new file? ************************************/
type response struct {
	Entity entity `json: "entity"`
}

type entity struct {
	Ports []int `json:"ports"`
}

func getPortsForApp(cliConn cliCommandRunner, guid string) ([]int, error) {
	appsEndpoint := fmt.Sprintf("/v2/apps/%s", guid)
	output, err := cliConn.CliCommandWithoutTerminalOutput("curl", appsEndpoint)
	if err != nil {
		return []int{}, err
	}
	joined := strings.Join(output, "")

	response := response{}
	err = json.Unmarshal([]byte(joined), &response)
	return response.Entity.Ports, err
}

func setPortsForApp(cliConn cliCommandRunner, guid string, ports []int) error {
	appsEndpoint := fmt.Sprintf("/v2/apps/%s", guid)

	newPortsEntity := entity{Ports: ports}
	portsBody, err := json.Marshal(newPortsEntity)
	if err != nil {
		return err
	}

	// TODO: do we need to wrap it in single quotes like we do in the term?
	wrappedPortsBody := fmt.Sprintf("'%s'", string(portsBody))
	_, err = cliConn.CliCommandWithoutTerminalOutput("curl", appsEndpoint, "-X", "PUT", "-d", wrappedPortsBody)
	return err
}

/************************************************************************/

func formatHost(r pluginmodels.GetApp_RouteSummary) string {
	host := r.Domain.Name
	if r.Host != "" {
		host = fmt.Sprintf("%s.%s", r.Host, host)
	}
	if r.Port != 0 {
		host = fmt.Sprintf("%s:%d", host, r.Port)
	}
	return host
}

func ensureHttpsPrefix(requestedRoute string) string {
	return "https://" + strings.Replace(requestedRoute, "https://", "", 1)
}

func ensureServiceAndBind(cliConn cliCommandRunner, appName, serviceProtocol, config string) error {
	serviceName := generateServiceName(serviceProtocol, config)
	exists, err := findExistingService(cliConn, serviceName)
	if err != nil {
		return err
	}

	if !exists {
		binding := serviceProtocol + "://" + config
		_, err = cliConn.CliCommandWithoutTerminalOutput("create-user-provided-service", serviceName, "-l", binding)
		if err != nil {
			return err
		}
	}

	_, err = cliConn.CliCommandWithoutTerminalOutput("bind-service", appName, serviceName)

	return err
}

func generateServiceName(serviceProtocol string, config string) string {
	cleanedConfig := sanitizeConfig(config)
	serviceName := serviceProtocol + "-" + cleanedConfig
	// Cloud Controller limits service name lengths:
	// see https://github.com/cloudfoundry/cloud_controller_ng/blob/master/vendor/errors/v2.yml#L231
	if len(serviceName) > 50 {
		hasher := sha1.New()
		hasher.Write([]byte(cleanedConfig))
		serviceName = serviceProtocol + "-" + strings.Trim(base64.URLEncoding.EncodeToString(hasher.Sum(nil)), "=")
	}
	return serviceName
}

func sanitizeConfig(config string) string {
	slashToDashes := strings.Replace(config, "/", "-", -1)
	removeColons := strings.Replace(slashToDashes, ":", "", -1)
	return strings.Trim(removeColons, "-")
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
