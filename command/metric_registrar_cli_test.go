package command_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "code.cloudfoundry.org/cli/plugin/models"
    "errors"
    "github.com/pivotal-cf/metric-registrar-cli/command"
)

var _ = Describe("CLI", func() {
    Context("RegisterLogFormat", func() {
        It("creates a service", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name"})
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "create-user-provided-service",
                "structured-format-format-name",
                "-l",
                "structured-format://format-name",
            )))
        })

        It("returns error if number of arguments is wrong", func() {
            cliConnection := newMockCliConnection()

            Expect(command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name", "some-garbage"})).To(HaveOccurred())
        })
    })

    Context("RegisterMetricsEndpoint", func() {
        It("creates a service", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, []string{"app-name", "endpoint"})
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "create-user-provided-service",
                "metrics-endpoint-endpoint",
                "-l",
                "metrics-endpoint://endpoint",
            )))
        })

        It("returns error if number of arguments is wrong", func() {
            cliConnection := newMockCliConnection()

            Expect(command.RegisterMetricsEndpoint(cliConnection, []string{"app-name", "endpoint", "some-garbage"})).To(HaveOccurred())
        })
    })

    Context("EnsureServiceAndBind", func() {
        It("creates a service and binds it to the application", func() {
            cliConnection := newMockCliConnection()

            err := command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "create-user-provided-service",
                "protocol-config",
                "-l",
                "protocol://config",
            )))

            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "bind-service",
                "app-name",
                "protocol-config",
            )))
        })

        It("doesn't create a service if service already present", func() {
            cliConnection := newMockCliConnection()
            cliConnection.serviceName = "protocol-config"

            err := command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("replaces slashes in the service name", func() {
            cliConnection := newMockCliConnection()

            err := command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "/v2/path/")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "create-user-provided-service",
                "protocol-v2-path",
                "-l",
                "protocol:///v2/path/",
            )))
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesError = errors.New("error")

            Expect(command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")).Should(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliCommandErrorCommand = "create-user-provided-service"

            Expect(command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")).Should(HaveOccurred())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if binding fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliCommandErrorCommand = "bind-service"

            Expect(command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")).Should(HaveOccurred())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
        })
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
