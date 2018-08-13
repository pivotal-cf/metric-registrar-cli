package command_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "cliplugin/command"
    "errors"
    "code.cloudfoundry.org/cli/plugin/models"
)

var _ = Describe("RegisterLogFormat", func() {
    It("creates a service and binds it to the application", func() {
        cliConnection := newMockCliConnection()

        command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name"})

        Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
            "create-user-provided-service",
            "structured-format-format-name",
            "-l",
            "structured-format://format-name",
        )))

        Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
            "bind-service",
            "app-name",
            "structured-format-format-name",
        )))
    })

    It("panics if number of arguments is wrong", func() {
        cliConnection := newMockCliConnection()

        Expect(func() {
            command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name", "some-garbage"})
        }).To(Panic())
    })

    It("Doesn't create a service if service already present", func() {
        cliConnection := newMockCliConnection()
        cliConnection.serviceName = "structured-format-format-name"

        command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name"})

        Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
        Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
    })

    It("panics if creating the service fails", func() {
        cliConnection := newMockCliConnection()
        cliConnection.getServicesError = errors.New("error")

        Expect(func() {
            command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name"})
        }).To(Panic())
        Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
    })

    It("panics if creating the service fails", func() {
        cliConnection := newMockCliConnection()
        cliConnection.cliCommandErrorCommand = "create-user-provided-service"

        Expect(func() {
            command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name"})
        }).To(Panic())

        Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
        Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
    })

    It("panics if binding fails", func() {
        cliConnection := newMockCliConnection()
        cliConnection.cliCommandErrorCommand = "bind-service"

        Expect(func() {
            command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name"})
        }).To(Panic())

        Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
        Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
    })

})

type mockCliConnection struct {
    cliCommandsCalled      chan []string
    serviceName            string
    cliCommandErrorCommand string
    getServicesError       error
}

func (c *mockCliConnection) GetServices() ([]plugin_models.GetServices_Model, error) {
    if c.serviceName != "" {
        return []plugin_models.GetServices_Model{{Name: c.serviceName}}, nil
    }
    return nil, c.getServicesError
}

func newMockCliConnection() *mockCliConnection {
    return &mockCliConnection{
        cliCommandsCalled: make(chan []string, 10),
    }
}

func (c *mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
    c.cliCommandsCalled <- args
    if args[0] == c.cliCommandErrorCommand {
        return nil, errors.New("error")
    }
    return nil, nil
}
