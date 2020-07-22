package command_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"github.com/pivotal-cf/metric-registrar-cli/registrations"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Suite")
}

type mockCliConnection struct {
	cliCommandsCalled chan []string
	cliErrorCommand   string

	getServicesResult []plugin_models.GetServices_Model
	getServicesError  error

	getAppResult plugin_models.GetAppModel
	getAppError  error

	getAppsResult []plugin_models.GetAppsModel
	getAppsError  error

	exposedPorts     []int
	getAppsInfoError error
	putAppsInfoError error
}

func (c *mockCliConnection) GetServices() ([]plugin_models.GetServices_Model, error) {
	return c.getServicesResult, c.getServicesError
}

func newMockCliConnection() *mockCliConnection {
	return &mockCliConnection{
		cliCommandsCalled: make(chan []string, 10),
		getAppResult: plugin_models.GetAppModel{
			Guid: "app-guid",
			Name: "app-name",
			Routes: []plugin_models.GetApp_RouteSummary{{
				Host: "app-host",
				Domain: plugin_models.GetApp_DomainFields{
					Name: "app-domain",
				},
				Path: "app-path",
			}},
		},
		exposedPorts: []int{8080},
	}
}

func (c *mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	c.cliCommandsCalled <- args

	// curl /v2/apps
	if args[0] == "curl" && strings.Contains(args[1], "/v2/apps/") {
		response := getFakeAppsInfoResponse(c.exposedPorts)
		output := strings.Split(response, "\n")

		if len(args) == 2 {
			return output, c.getAppsInfoError
		} else {
			return output, c.putAppsInfoError
		}
	}

	if args[0] == c.cliErrorCommand {
		return nil, errors.New("error")
	}
	return nil, nil
}

func (c *mockCliConnection) GetApp(string) (plugin_models.GetAppModel, error) {
	return c.getAppResult, c.getAppError
}

func (c *mockCliConnection) GetApps() ([]plugin_models.GetAppsModel, error) {
	return c.getAppsResult, c.getAppsError
}

type mockRegistrationFetcher struct {
	registrations map[string][]registrations.Registration
	fetchError    error
}

func newMockRegistrationFetcher() *mockRegistrationFetcher {
	return &mockRegistrationFetcher{
		registrations: map[string][]registrations.Registration{},
	}
}

func (f *mockRegistrationFetcher) Fetch(appGuid, registrationType string) ([]registrations.Registration, error) {
	Expect(appGuid).To(Equal("app-guid"))

	reg := make([]registrations.Registration, 0)
	for _, r := range f.registrations["app-guid"] {
		if r.Type == registrationType {
			reg = append(reg, r)
		}
	}

	return reg, f.fetchError
}

func (f *mockRegistrationFetcher) FetchAll(registrationType string) (map[string][]registrations.Registration, error) {
	return f.registrations, f.fetchError
}

func getFakeAppsInfoResponse(ports []int) string {
	portsString := strings.Replace(fmt.Sprint(ports), " ", ",", -1)
	return fmt.Sprintf(fakeAppsInfoResponseTemplate, portsString)
}

const (
	fakeAppsInfoResponseTemplate = `{
   "metadata": {
      "guid": "419aa316-c70b-40f9-b38b-c99e693e7620",
      "url": "/v2/apps/419aa316-c70b-40f9-b38b-c99e693e7620",
      "created_at": "2020-07-21T22:09:59Z",
      "updated_at": "2020-07-21T22:17:30Z"
   },
   "entity": {
      "name": "go-app-with-metrics",
      "production": false,
      "space_guid": "eefee588-c08a-4fde-b0e3-8b2a6309e365",
      "stack_guid": "a3fd69f2-b350-4515-86a9-7fab406e8544",
      "buildpack": null,
      "detected_buildpack": "go",
      "detected_buildpack_guid": "dc517d4d-694b-4e8a-a014-42b5a55b9d9f",
      "environment_json": {},
      "memory": 1024,
      "instances": 2,
      "disk_quota": 1024,
      "state": "STARTED",
      "version": "bb68c933-3ede-452c-9a3a-709f264d43ed",
      "command": null,
      "console": false,
      "debug": null,
      "staging_task_id": "0c11278f-beb2-4681-8518-44c1b7a5545a",
      "package_state": "STAGED",
      "health_check_type": "process",
      "health_check_timeout": null,
      "health_check_http_endpoint": "",
      "staging_failed_reason": null,
      "staging_failed_description": null,
      "diego": true,
      "docker_image": null,
      "docker_credentials": {
         "username": null,
         "password": null
      },
      "package_updated_at": "2020-07-21T22:09:59Z",
      "detected_start_command": "./bin/go-app-with-metrics",
      "enable_ssh": true,
      "ports": %s,
      "space_url": "/v2/spaces/eefee588-c08a-4fde-b0e3-8b2a6309e365",
      "stack_url": "/v2/stacks/a3fd69f2-b350-4515-86a9-7fab406e8544",
      "routes_url": "/v2/apps/419aa316-c70b-40f9-b38b-c99e693e7620/routes",
      "events_url": "/v2/apps/419aa316-c70b-40f9-b38b-c99e693e7620/events",
      "service_bindings_url": "/v2/apps/419aa316-c70b-40f9-b38b-c99e693e7620/service_bindings",
      "route_mappings_url": "/v2/apps/419aa316-c70b-40f9-b38b-c99e693e7620/route_mappings"
   }
}
`
)
