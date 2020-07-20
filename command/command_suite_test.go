package command_test

import (
	"errors"
	"testing"

	"github.com/pivotal-cf/metric-registrar-cli/registrations"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
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
	}
}

func (c *mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	if args[0] == "curl" {
		return []string{}, nil
	}

	c.cliCommandsCalled <- args
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
	return f.registrations["app-guid"], f.fetchError
}

func (f *mockRegistrationFetcher) FetchAll(registrationType string) (map[string][]registrations.Registration, error) {
	return f.registrations, f.fetchError
}
