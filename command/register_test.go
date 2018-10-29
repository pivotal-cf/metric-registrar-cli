package command_test

import (
    "code.cloudfoundry.org/cli/plugin/models"
    "errors"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/pivotal-cf/metric-registrar-cli/command"
)

var _ = Describe("Register", func() {
    Context("RegisterLogFormat", func() {
        It("creates a service", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterLogFormat(cliConnection, "app-name", "format-name")
            Expect(err).ToNot(HaveOccurred())
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

        It("doesn't create a service if service already present", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesResult = []plugin_models.GetServices_Model{
                {Name: "structured-format-config"},
            }

            err := command.RegisterLogFormat(cliConnection, "app-name", "config")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if getting the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesError = errors.New("error")

            Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "create-user-provided-service"

            Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if binding fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "bind-service"

            Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
        })
    })

    Context("RegisterMetricsEndpoint", func() {
        It("creates a service", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "endpoint")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "create-user-provided-service",
                "metrics-endpoint-endpoint",
                "-l",
                "metrics-endpoint://endpoint",
            )))

            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "bind-service",
                "app-name",
                "metrics-endpoint-endpoint",
            )))
        })

        It("doesn't create a service if service already present", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesResult = []plugin_models.GetServices_Model{
                {Name: "metrics-endpoint-config"},
            }

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "config")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("replaces slashes in the service name", func() {
            cliConnection := newMockCliConnection()

            err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/v2/path/")
            Expect(err).ToNot(HaveOccurred())
            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "create-user-provided-service",
                "metrics-endpoint-v2-path",
                "-l",
                "metrics-endpoint:///v2/path/",
            )))
        })

        It("returns error if getting the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesError = errors.New("error")

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "config")).ToNot(Succeed())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "create-user-provided-service"

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if binding fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "bind-service"

            Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
        })
    })
})
