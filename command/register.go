package command

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pivotal-cf/metric-registrar-cli/ports"

	pluginmodels "code.cloudfoundry.org/cli/plugin/models"
)

const (
	secureWarning = "Using a secure endpoint by specifying --internal-port is preferred, but not supported in all versions. See documentation for further details https://docs.pivotal.io/platform/application-service/2-10/metric-registrar/using.html"
)

func RegisterLogFormat(cliConn cliCommandRunner, appName, logFormat string) error {
	return ensureServiceAndBind(cliConn, appName, structuredFormat, logFormat)
}

func RegisterMetricsEndpoint(cliConn cliCommandRunner, appName, route, internalPort string, insecure bool) error {
	// validate flags
	if internalPort == "" && !insecure {
		return fmt.Errorf("need to pass either --internal-port or --insecure")
	}

	app, err := cliConn.GetApp(appName)
	if err != nil {
		return err
	}

	validRoute, err := validateRouteForApp(route, app, !insecure)
	if err != nil {
		return err
	}

	fmt.Printf(secureWarning)
	serviceProtocol := metricsEndpoint
	if !insecure {
		route = ":" + internalPort + validRoute.Path
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

func validateRouteForApp(requestedRoute string, app pluginmodels.GetAppModel, secure bool) (url.URL, error) {
	// if they just provide a relative path, we can associate it with the app
	if requestedRoute[0] == '/' {
		return url.URL{Path: requestedRoute}, nil
	}

	requested, err := url.Parse(ensureHttpsPrefix(requestedRoute))
	if err != nil {
		return url.URL{}, fmt.Errorf("unable to parse requested route: %s", err)
	}

	if secure && requested.Host != "" {
		return url.URL{}, fmt.Errorf("cannot provide hostname with --internal-port. provided: '%s'", requested.Host)
	}

	for _, r := range app.Routes {
		var host string
		host = formatHost(r)
		route := url.URL{
			Host: host,
			Path: "/" + strings.TrimPrefix(r.Path, "/"),
		}

		if requested.Host == route.Host && strings.HasPrefix(requested.Path, route.Path) {
			return route, nil
		}

	}
	return url.URL{}, fmt.Errorf("route '%s' is not bound to app '%s'", requestedRoute, app.Name)
}

func exposePortForApp(cliConn cliCommandRunner, guid string, port int) error {
	existingPorts, err := ports.GetPortsForApp(cliConn, guid)
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
	return ports.SetPortsForApp(cliConn, guid, newPorts)
}

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
