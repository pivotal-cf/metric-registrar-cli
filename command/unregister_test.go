package command_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pivotal-cf/metric-registrar-cli/command"
	"github.com/pivotal-cf/metric-registrar-cli/registrations"

	. "github.com/benjamintf1/unmarshalledmatchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unregister", func() {
	Context("UnregisterLogFormat", func() {
		It("unbinds app from all log services", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "structured-format",
					Config:           "json",
					NumberOfBindings: 2,
				},
				{
					Name:             "service2",
					Type:             "structured-format",
					Config:           "json",
					NumberOfBindings: 2,
				},
			}

			err := command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")
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
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "structured-format",
					Config:           "json",
					NumberOfBindings: 1,
				},
			}

			err := command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")
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

		It("only unbinds specified service if format is set", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "structured-format",
					Config:           "json",
					NumberOfBindings: 2,
				},
				{
					Name:             "service2",
					Type:             "structured-format",
					Config:           "not-this-one",
					NumberOfBindings: 2,
				},
			}

			err := command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "json")
			Expect(err).ToNot(HaveOccurred())

			Eventually(cliConnection.cliCommandsCalled).Should(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"service1",
			)))

			Expect(cliConnection.cliCommandsCalled).To(BeEmpty())
		})

		It("doesn't unbind services if registration fetcher doesn't find any", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = nil

			err := command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(cliConnection.cliCommandsCalled).ToNot(Receive())
		})

		It("returns error if getting app info fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.getAppError = errors.New("expected")
			registrationFetcher := newMockRegistrationFetcher()

			Expect(command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")).ToNot(Succeed())
		})

		It("returns error if unbinding service fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "unbind-service"
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "structured-format",
					Config:           "json",
					NumberOfBindings: 1,
				},
			}

			Expect(command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")).ToNot(Succeed())
		})

		It("returns error if deleting service fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "delete-service"
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "structured-format",
					Config:           "json",
					NumberOfBindings: 1,
				},
			}

			Expect(command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")).ToNot(Succeed())
		})

		It("returns an error if registration fetcher returns an error", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.fetchError = errors.New("expected")

			Expect(command.UnregisterLogFormat(registrationFetcher, cliConnection, "app-name", "")).ToNot(Succeed())
		})
	})

	Context("UnregisterMetricsEndpoint", func() {
		It("unbinds app from all metrics endpoints", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "secure-endpoint-2112-metrics",
					Type:             "secure-endpoint",
					Config:           ":2112/metrics",
					NumberOfBindings: 2,
				},
				{
					Name:             "service2",
					Type:             "metrics-endpoint",
					Config:           ":9090/metrics",
					NumberOfBindings: 2,
				},
			}

			err := command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")
			Expect(err).ToNot(HaveOccurred())

			Eventually(cliConnection.cliCommandsCalled).Should(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"service2",
			)))

			Eventually(cliConnection.cliCommandsCalled).Should(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"secure-endpoint-2112-metrics",
			)))
		})

		It("removes exposed ports", func() {
			cliConnection := newMockCliConnection()
			cliConnection.exposedPorts = []int{1234, 2112}

			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "metrics-endpoint",
					Config:           ":2112/metrics",
					NumberOfBindings: 1,
				},
			}
			err := command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")
			Expect(err).ToNot(HaveOccurred())
			expectToReceivePutCurlForAppAndPort(cliConnection.cliCommandsCalled, "app-guid", []string{"1234"})
		})

		It("deletes service if no more apps bound", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "metrics-endpoint",
					Config:           ":9090/metrics",
					NumberOfBindings: 1,
				},
			}

			err := command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")
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

		It("only unbinds specified service if path and port are set", func() {
			cliConnection := newMockCliConnection()
			cliConnection.exposedPorts = []int{8080, 9090}
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "secure-endpoint",
					Config:           ":9090/metrics",
					NumberOfBindings: 2,
				},
				{
					Name:             "service2",
					Type:             "secure-endpoint",
					Config:           ":9090/not-this-one",
					NumberOfBindings: 2,
				},
			}

			err := command.UnregisterMetricsEndpoint(
				registrationFetcher,
				cliConnection,
				"app-name",
				"/metrics",
				"9090",
			)
			Expect(err).ToNot(HaveOccurred())

			Expect(cliConnection.cliCommandsCalled).Should(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"service1",
			)))

			Expect(cliConnection.cliCommandsCalled).ShouldNot(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"service2",
			)))
			expectToReceivePutCurlForAppAndPort(cliConnection.cliCommandsCalled, "app-guid", []string{"8080", "9090"})
		})

		It("only unbinds specified service if path is set", func() {
			cliConnection := newMockCliConnection()
			cliConnection.exposedPorts = []int{8080, 9090}
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "metrics-endpoint",
					Config:           ":9090/metrics",
					NumberOfBindings: 2,
				},
				{
					Name:             "service2",
					Type:             "metrics-endpoint",
					Config:           ":9090/not-this-one",
					NumberOfBindings: 2,
				},
			}

			err := command.UnregisterMetricsEndpoint(
				registrationFetcher,
				cliConnection,
				"app-name",
				":9090/metrics",
				"",
			)
			Expect(err).ToNot(HaveOccurred())

			Expect(cliConnection.cliCommandsCalled).Should(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"service1",
			)))

			Expect(cliConnection.cliCommandsCalled).ShouldNot(Receive(ConsistOf(
				"unbind-service",
				"app-name",
				"service2",
			)))

			expectToReceivePutCurlForAppAndPort(cliConnection.cliCommandsCalled, "app-guid", []string{"8080", "9090"})
		})

		It("doesn't unbind services if registration fetcher doesn't find any", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = nil

			err := command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(cliConnection.cliCommandsCalled).ShouldNot(Receive(ContainElement("unbind-service")))
		})

		It("returns error if getting app info fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.getAppError = errors.New("expected")
			registrationFetcher := newMockRegistrationFetcher()

			Expect(command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")).ToNot(Succeed())
		})

		It("returns error if unbinding service fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "unbind-service"
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "metrics-endpoint",
					Config:           "/metrics",
					NumberOfBindings: 1,
				},
			}

			Expect(command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")).ToNot(Succeed())
		})

		It("returns error if deleting service fails", func() {
			cliConnection := newMockCliConnection()
			cliConnection.cliErrorCommand = "delete-service"
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.registrations["app-guid"] = []registrations.Registration{
				{
					Name:             "service1",
					Type:             "metrics-endpoint",
					Config:           "/metrics",
					NumberOfBindings: 1,
				},
			}

			Expect(command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")).ToNot(Succeed())
		})

		It("returns an error if registration fetcher returns an error", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			registrationFetcher.fetchError = errors.New("expected")

			Expect(command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "", "")).ToNot(Succeed())
		})

		It("returns an error if unregistering the port returns an error", func() {
			cliConnection := newMockCliConnection()
			registrationFetcher := newMockRegistrationFetcher()
			cliConnection.getAppsInfoError = errors.New("cf doesn't want to speak to you rn")

			Expect(command.UnregisterMetricsEndpoint(registrationFetcher, cliConnection, "app-name", "2112", "")).ToNot(Succeed())
		})
	})
})

func expectToReceivePutCurlForAppAndPort(called chan []string, appGuid string, ports []string) {
	Eventually(called).Should(Receive(And(
		matchCurl(
			fmt.Sprintf("/v2/apps/%s", appGuid),
			"-X",
			"PUT",
			"-d",
		), ContainElement(
			WithTransform(removeSingleQuotes, MatchUnorderedJSON(
				fmt.Sprintf("{\"ports\":[%s]}", strings.Join(ports, ",")),
			)),
		),
	)))
}

func removeSingleQuotes(s string) interface{} {
	return strings.ReplaceAll(s, "'", "")
}
