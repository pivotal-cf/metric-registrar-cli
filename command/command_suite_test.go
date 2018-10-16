package command_test

import (
    "code.cloudfoundry.org/cli/plugin/models"
    "errors"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"
    "testing"

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

    services         []plugin_models.GetServices_Model
    getServicesError error

    getAppError  error
    getAppResult plugin_models.GetAppModel
}

func (c *mockCliConnection) GetServices() ([]plugin_models.GetServices_Model, error) {
    return c.services, c.getServicesError
}

func newMockCliConnection() *mockCliConnection {
    return &mockCliConnection{
        cliCommandsCalled: make(chan []string, 10),
        getAppResult: plugin_models.GetAppModel{
            Guid: "app-guid",
        },
    }
}

func (c *mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
    if args[0] == "curl" {
        return []string{

        }, nil
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

type mockRegistrationFetcher struct {
    registrations []registrations.Registration
    fetchError    error
}

func newMockRegistrationFetcher() *mockRegistrationFetcher {
    return &mockRegistrationFetcher{

    }
}

func (f *mockRegistrationFetcher) Fetch(appGuid, registrationType string) ([]registrations.Registration, error) {
    Expect(appGuid).To(Equal("app-guid"))
    return f.registrations, f.fetchError
}
