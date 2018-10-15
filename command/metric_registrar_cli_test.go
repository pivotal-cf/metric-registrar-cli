package command_test

import (
    "code.cloudfoundry.org/cli/plugin"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/onsi/gomega/gstruct"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"

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
            cliConnection.services = []plugin_models.GetServices_Model{
                {Name: "protocol-config"},
            }

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

        It("returns error if getting the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.getServicesError = errors.New("error")

            Expect(command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")).ToNot(Succeed())
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if creating the service fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "create-user-provided-service"

            Expect(command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })

        It("returns error if binding fails", func() {
            cliConnection := newMockCliConnection()
            cliConnection.cliErrorCommand = "bind-service"

            Expect(command.EnsureServiceAndBind(cliConnection, "app-name", "protocol", "config")).ToNot(Succeed())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("create-user-provided-service")))
            Expect(cliConnection.cliCommandsCalled).To(Receive(ContainElement("bind-service")))
        })
    })

    Context("metadata", func() {
        It("outputs the correct metadata", func() {
            meta := command.MetricRegistrarCli{
                Major: 1,
                Minor: 2,
                Patch: 3,
            }.GetMetadata()

            Expect(meta.Name).Should(Equal("metric-registrar"))
            Expect(meta.Version).Should(Equal(plugin.VersionType{Major: 1, Minor: 2, Build: 3}))
            Expect(meta.Commands).To(ConsistOf(
                gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("register-log-format")}),
                gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("register-metrics-endpoint")}),
            ))
        })
    })

    Context("UnregisterLogFormat", func() {
        It("unbinds app from all log services", func() {
            cliConnection := newMockCliConnection()
            cliConnection.services = []plugin_models.GetServices_Model{
                {Name: "service1", ApplicationNames: []string{"app-name", "other-app-name"}},
                {Name: "service2", ApplicationNames: []string{"app-name", "other-app-name"}},
            }

            registrationFetcher := newMockRegistrationFetcher()

            err := command.UnregisterLogFormat(registrationFetcher, cliConnection, []string{"app-name"})
            Expect(err).ToNot(HaveOccurred())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "unbind-service",
                "app-name",
                "service1",
            )))

            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "unbind-service",
                "app-name",
                "service2",
            )))
        })

        It("deletes service if no more apps bound", func() {
            cliConnection := newMockCliConnection()
            cliConnection.services = []plugin_models.GetServices_Model{
                {Name: "service1", ApplicationNames: []string{"app-name"}},
            }

            registrationFetcher := newMockRegistrationFetcher()
            err := command.UnregisterLogFormat(registrationFetcher, cliConnection, []string{"app-name"})
            Expect(err).ToNot(HaveOccurred())

            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "unbind-service",
                "app-name",
                "service1",
            )))

            Expect(cliConnection.cliCommandsCalled).To(Receive(ConsistOf(
                "delete-service",
                "service1",
                "-f",
            )))
        })

        It("doesn't unbind non-log-format services", func() {
            cliConnection := newMockCliConnection()
            cliConnection.services = []plugin_models.GetServices_Model{
                {
                    Guid:        "",
                    Name:        "service1",
                    ServicePlan: plugin_models.GetServices_ServicePlan{},
                    Service: plugin_models.GetServices_ServiceFields{
                        Name: "",
                    },
                    LastOperation:    plugin_models.GetServices_LastOperation{},
                    ApplicationNames: []string{"app-name"},
                    IsUserProvided:   false,
                },
            }

            registrationFetcher := newMockRegistrationFetcher()
            err := command.UnregisterLogFormat(registrationFetcher, cliConnection, []string{"app-name"})
            Expect(err).ToNot(HaveOccurred())

            Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
        })
        XIt("doesn't attempt to unbind services that aren't bound to this app")
        XIt("returns error if number of arguments is wrong", func() {
            cliConnection := newMockCliConnection()

            Expect(command.RegisterLogFormat(cliConnection, []string{"app-name", "format-name", "some-garbage"})).ToNot(Succeed())
        })
    })

    XContext("UnregisterMetricsEndpoint", func() {
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

            Expect(command.RegisterMetricsEndpoint(cliConnection, []string{"app-name", "endpoint", "some-garbage"})).ToNot(Succeed())
        })
    })
})

type mockCliConnection struct {
    cliCommandsCalled chan []string
    services          []plugin_models.GetServices_Model
    cliErrorCommand   string
    getServicesError  error
}

func (c *mockCliConnection) GetServices() ([]plugin_models.GetServices_Model, error) {
    return c.services, c.getServicesError
}

func newMockCliConnection() *mockCliConnection {
    return &mockCliConnection{
        cliCommandsCalled: make(chan []string, 10),
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
    return plugin_models.GetAppModel{
        Guid: "app-guid",
    }, nil //TODO err
}

type mockRegistrationFetcher struct {
}

func newMockRegistrationFetcher() *mockRegistrationFetcher {
    return &mockRegistrationFetcher{

    }
}

func (f *mockRegistrationFetcher) Fetch(appGuid, registrationType string) ([]registrations.Registration, error) {
    Expect(appGuid).To(Equal("app-guid"))
    return nil, nil
}
