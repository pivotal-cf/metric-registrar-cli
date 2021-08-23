package command_test

import (
	"errors"
	"fmt"
	"strings"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"github.com/pivotal-cf/metric-registrar-cli/command"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Register", func() {
	Context("RegisterLogFormat", func() {
		It("creates a service", func() {
			cliConnection := newMockCliConnection()

			err := command.RegisterLogFormat(cliConnection, "app-name", "format-name")
			Expect(err).ToNot(HaveOccurred())
			Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
				"structured-format-format-name",
				"-l",
				"structured-format://format-name",
			))

			Expect(cliConnection.cliCommandsCalled).To(receiveBindService(
				"app-name",
				"structured-format-format-name",
			))
		})

		It("doesn't create a service if service already present", func() {
			cliConnection := newMockCliConnection()
			cliConnection.getServicesResult = []plugin_models.GetServices_Model{
				{Name: "structured-format-config"},
			}

			err := command.RegisterLogFormat(cliConnection, "app-name", "config")
			Expect(err).ToNot(HaveOccurred())
			Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
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

			Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService())
			Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
		})

		It("returns error if binding fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "bind-service"

			Expect(command.RegisterLogFormat(cliConnection, "app-name", "config")).ToNot(Succeed())

			Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService())
			Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
		})
	})

	Context("RegisterMetricsEndpoint", func() {
		It("fails if neither --internal-port or --insecure is passed", func() {
			cliConnection := newMockCliConnection()

			err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "", false)
			Expect(err).To(HaveOccurred())
		})

		It("does not use service names longer than 50 characters", func() {
			cliConnection := newMockCliConnection()

			err := command.RegisterMetricsEndpoint(cliConnection, "very-long-app-name-with-many-characters", "/metrics", "8091", false)
			Expect(err).ToNot(HaveOccurred())

			Eventually(cliConnection.cliCommandsCalled).Should(Receive(ConsistOf(
				"create-user-provided-service",
				WithTransform(func(s string) int { return len(s) }, BeNumerically("<=", 50)),
				"-l",
				HavePrefix("secure-endpoint://"),
			)))

			Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
		})

		It("doesn't create a service if service already present", func() {
			cliConnection := newMockCliConnection()
			cliConnection.getServicesResult = []plugin_models.GetServices_Model{
				{Name: "secure-endpoint-8091-metrics"},
			}

			err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "8091", false)
			Expect(err).ToNot(HaveOccurred())

			// Ignore the getting and setting ports call
			<-cliConnection.cliCommandsCalled
			<-cliConnection.cliCommandsCalled

			var received []string
			Expect(cliConnection.cliCommandsCalled).To(Receive(&received))
			Expect(received).ToNot(matchCreateUserProvidedService())
			Expect(received).To(matchBindService())
		})

		It("replaces slashes in the service name", func() {
			cliConnection := newMockCliConnection()

			err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/v2/path/", "8091", false)
			Expect(err).ToNot(HaveOccurred())
			Eventually(cliConnection.cliCommandsCalled).Should(receiveCreateUserProvidedService(
				"secure-endpoint-8091-v2-path",
				"-l",
				"secure-endpoint://:8091/v2/path/",
			))
		})

		It("returns error if getting the service fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.getServicesError = errors.New("error")

			Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "", true)).ToNot(Succeed())
			Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
		})

		It("returns error if creating the service fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "create-user-provided-service"

			Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "8091", false)).ToNot(Succeed())

			Eventually(cliConnection.cliCommandsCalled).Should(receiveCreateUserProvidedService())
			Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
		})

		It("returns error if binding fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "bind-service"

			Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "8091", false)).ToNot(Succeed())

			Eventually(cliConnection.cliCommandsCalled).Should(receiveCreateUserProvidedService())
			Expect(cliConnection.cliCommandsCalled).To(receiveBindService())
		})

		It("returns error if getting the app fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.getAppError = errors.New("error")

			Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "app-host.app-domain/app-path/metrics", "8091", false)).ToNot(Succeed())
			Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
		})

		It("returns an error if parsing the route fails", func() {
			cliConnection := newMockCliConnection()

			err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "#$%#$%#", "8091", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("unable to parse requested route:"))
			Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
		})

		Context("when secure", func() {
			It("errors when domain is passed", func() {
				cliConnection := newMockCliConnection()

				err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "app-host.app-domain/app-path/metrics", "8091", false)
				Expect(err).To(MatchError("cannot provide hostname with --internal-port. provided: 'app-host.app-domain'"))
			})

			It("creates a service given a path", func() {
				cliConnection := newMockCliConnection()

				err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "1234", false)
				Expect(err).ToNot(HaveOccurred())

				Eventually(cliConnection.cliCommandsCalled).Should(receiveCreateUserProvidedService(
					"secure-endpoint-1234-metrics",
					"-l",
					"secure-endpoint://:1234/metrics",
				))

				Expect(cliConnection.cliCommandsCalled).To(receiveBindService(
					"app-name",
					"secure-endpoint-1234-metrics",
				))
			})

			It("exposes the internal port automatically and preserves existing ports", func() {
				cliConnection := newMockCliConnection()
				cliConnection.exposedPorts = []int{1234}

				Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/v2/metrics", "2112", false)).To(Succeed())
				expectToReceiveCurlForAppAndPort(cliConnection.cliCommandsCalled, "app-guid", []string{"1234", "2112"})
			})

			It("returns error if getting existing ports fails", func() {
				cliConnection := newMockCliConnection()
				cliConnection.getAppsInfoError = errors.New("failed to fetch apps info")

				Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/v2/metrics", "2112", false)).ToNot(Succeed())
			})

			It("returns error if setting port fails", func() {
				cliConnection := newMockCliConnection()
				cliConnection.putAppsInfoError = errors.New("failed to put apps info")

				Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "/v2/metrics", "2112", false)).ToNot(Succeed())
			})
		})

		Context("when insecure", func() {
			It("creates a metrics-endpoint", func() {
				cliConnection := newMockCliConnection()

				err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "", true)
				Expect(err).ToNot(HaveOccurred())

				Eventually(cliConnection.cliCommandsCalled).Should(receiveCreateUserProvidedService(
					"metrics-endpoint-metrics",
					"-l",
					"metrics-endpoint:///metrics",
				))

				Expect(cliConnection.cliCommandsCalled).To(receiveBindService(
					"app-name",
					"metrics-endpoint-metrics",
				))
			})

			It("creates a service given a path", func() {
				cliConnection := newMockCliConnection()

				err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "/metrics", "", true)
				Expect(err).ToNot(HaveOccurred())
				Eventually(cliConnection.cliCommandsCalled).Should(receiveCreateUserProvidedService(
					"metrics-endpoint-metrics",
					"-l",
					"metrics-endpoint:///metrics",
				))

				Expect(cliConnection.cliCommandsCalled).To(receiveBindService(
					"app-name",
					"metrics-endpoint-metrics",
				))
			})

			It("checks the route when domain is passed", func() {
				cliConnection := newMockCliConnection()
				err := command.RegisterMetricsEndpoint(cliConnection, "app-name", "not-app-host.app-domain/app-path/metrics", "", true)
				Expect(err).To(MatchError("route 'not-app-host.app-domain/app-path/metrics' is not bound to app 'app-name'"))
			})

			It("checks the route when domain is passed correctly", func() {
				cliConnection := newMockCliConnection()
				Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "app-host.app-domain/app-path", "", true)).To(Succeed())
				Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
					"metrics-endpoint-app-host.app-domain-app-path",
					"-l",
					"metrics-endpoint://app-host.app-domain/app-path",
				))
			})

			It("checks the route when app route starts with //", func() {
				cliConnection := newMockCliConnection()

				cliConnection.getAppResult.Routes = []plugin_models.GetApp_RouteSummary{{
					Host: "app-host",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "app-domain",
					},
					Path: "/app-path",
				}}
				Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "app-host.app-domain/app-path", "", true)).To(Succeed())
				Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
					"metrics-endpoint-app-host.app-domain-app-path",
					"-l",
					"metrics-endpoint://app-host.app-domain/app-path",
				))
			})

			It("parses routes without hosts correctly", func() {
				cliConnection := newMockCliConnection()

				cliConnection.getAppResult.Routes = []plugin_models.GetApp_RouteSummary{{
					Host: "",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "tcp.app-domain",
					},
				}}

				Expect(command.RegisterMetricsEndpoint(cliConnection, "app-name", "tcp.app-domain/v2/path/", "", true)).To(Succeed())
				Expect(cliConnection.cliCommandsCalled).To(receiveCreateUserProvidedService(
					"metrics-endpoint-tcp.app-domain-v2-path",
					"-l",
					"metrics-endpoint://tcp.app-domain/v2/path/",
				))
			})
		})

	})
})

func expectToReceiveCupsArgs(called chan []string) (string, string) {
	var args []string
	Eventually(called).Should(Receive(&args))
	Expect(args).To(HaveLen(4))
	Expect(args[0]).To(Equal("create-user-provided-service"))
	Expect(args[2]).To(Equal("-l"))
	return args[1], args[3]
}

func matchCreateUserProvidedService(args ...string) types.GomegaMatcher {
	if len(args) == 0 {
		return ContainElement("create-user-provided-service")
	}

	return Equal(append([]string{"create-user-provided-service"}, args...))
}

func receiveCreateUserProvidedService(args ...string) types.GomegaMatcher {
	return Receive(matchCreateUserProvidedService(args...))
}

func matchCurl(args ...string) types.GomegaMatcher {
	if len(args) == 0 {
		return ContainElement("curl")
	}

	return ContainElements(append([]string{"curl"}, args...))
}

func expectToReceiveCurlForAppAndPort(called chan []string, appGuid string, ports []string) {
	Eventually(called).Should(Receive(matchCurl(
		fmt.Sprintf("/v2/apps/%s", appGuid),
		"-X",
		"PUT",
		"-d",
		fmt.Sprintf("'{\"ports\":[%s]}'", strings.Join(ports, ",")),
	)))
}

func matchBindService(args ...string) types.GomegaMatcher {
	if len(args) == 0 {
		return ContainElement("bind-service")
	}
	return Equal(append([]string{"bind-service"}, args...))
}

func receiveBindService(args ...string) types.GomegaMatcher {
	return Receive(matchBindService(args...))
}
