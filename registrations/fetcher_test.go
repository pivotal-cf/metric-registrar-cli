package registrations_test

import (
    "code.cloudfoundry.org/cli/plugin/models"
    "errors"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/ginkgo/extensions/table"
    . "github.com/onsi/gomega"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"
    "strings"
)

var _ = Describe("Fetcher", func() {
    It("Fetches registrations", func() {
        cliConn := newMockCliConnection()
        fetcher := registrations.NewFetcher(cliConn)

        s, err := fetcher.Fetch("app-guid", "structured-format")
        Expect(err).ToNot(HaveOccurred())
        Expect(s).To(ConsistOf(registrations.Registration{
            Name:             "structured-format-service",
            Type:             "structured-format",
            Config:           "json",
            NumberOfBindings: 2,
        }))
    })

    It("handles paging", func() {
        cliConn := newMockCliConnection()
        fetcher := registrations.NewFetcher(cliConn)

        cliConn.curlResponses["user_provided_service_instances"] = []string{validServicesPage0, validServices}
        cliConn.curlResponses["service_bindings"] = []string{
            validBindingsPage0, validBindings,
            validBindings,
        }

        s, err := fetcher.Fetch("app-guid", "structured-format")
        Expect(err).ToNot(HaveOccurred())
        Expect(s).To(ConsistOf(
            registrations.Registration{
                Name:             "structured-format-service-0",
                Type:             "structured-format",
                Config:           "json",
                NumberOfBindings: 3,
            },
            registrations.Registration{
                Name:             "structured-format-service",
                Type:             "structured-format",
                Config:           "json",
                NumberOfBindings: 2,
            },
        ))
    })

    DescribeTable("errors", func(modify func(*mockCliConnection)) {
        cliConn := newMockCliConnection()
        modify(cliConn)
        fetcher := registrations.NewFetcher(cliConn)

        _, err := fetcher.Fetch("app-guid", "structured-format")
        Expect(err).To(HaveOccurred())
    },
        Entry("getting space fails", func(cliConn *mockCliConnection) {
            cliConn.getCurrentSpaceError = errors.New("expected")
        }),
        Entry("getting service instances fails", func(cliConn *mockCliConnection) {
            cliConn.curlErrors["user_provided_service_instances"] = errors.New("expected")
        }),
        Entry("getting service instances returns invalid JSON", func(cliConn *mockCliConnection) {
            cliConn.curlResponses["user_provided_service_instances"] = []string{`{invalid]`}
        }),

        Entry("getting service bindings fails", func(cliConn *mockCliConnection) {
            cliConn.curlErrors["service_bindings"] = errors.New("expected")
        }),
        Entry("getting service bindings returns invalid JSON", func(cliConn *mockCliConnection) {
            cliConn.curlResponses["service_bindings"] = []string{`{invalid]`}
        }),
    )
})

type mockCliConnection struct {
    getCurrentSpaceError error
    curlResponses        map[string][]string
    curlErrors           map[string]error
}

func newMockCliConnection() *mockCliConnection {
    return &mockCliConnection{
        curlResponses: map[string][]string{
            "user_provided_service_instances": {validServices},
            "service_bindings":                {validBindings},
        },
        curlErrors: map[string]error{},
    }
}

func (c *mockCliConnection) GetCurrentSpace() (plugin_models.Space, error) {
    return plugin_models.Space{
        SpaceFields: plugin_models.SpaceFields{
            Guid: "space-guid",
        },
    }, c.getCurrentSpaceError
}

func (c *mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
    Expect(args[0]).To(Equal("curl"))
    pathWithQuery := args[1]

    parts := strings.Split(strings.Split(pathWithQuery, "?")[0], "/")
    resource := parts[len(parts)-1]

    switch resource {
    case "user_provided_service_instances":
        Expect(pathWithQuery).To(Equal("/v2/user_provided_service_instances?q=space_guid:space-guid"))
    case "service_bindings":
        if pathWithQuery != "/v2/user_provided_service_instances/guid/service_bindings" {
            return strings.Split(emptyBindings, "\n"), c.curlErrors[resource]
        }
    }

    resp := c.curlResponses[resource][0]
    c.curlResponses[resource] = c.curlResponses[resource][1:]

    return strings.Split(resp, "\n"), c.curlErrors[resource]
}

const (
    validServices = `{
  "next_url": null,
  "resources": [
    {
      "entity": {
        "name": "structured-format-service",
        "syslog_drain_url": "structured-format://json",
        "service_bindings_url": "/v2/user_provided_service_instances/guid/service_bindings"
      }
    },
    {
      "entity": {
        "name": "unbound-structured-format-service",
        "syslog_drain_url": "structured-format://json",
        "service_bindings_url": "/empty/service_bindings"
      }
    },
    {
      "entity": {
        "name": "other-valid-service",
        "syslog_drain_url": "not-structured-format://json",
        "service_bindings_url": "/empty/service_bindings"
      }
    }
  ]
}`
    validServicesPage0 = `{
  "next_url": "/v2/user_provided_service_instances?q=space_guid:space-guid",
  "resources": [
    {
      "entity": {
        "name": "structured-format-service-0",
        "syslog_drain_url": "structured-format://json",
        "service_bindings_url": "/v2/user_provided_service_instances/guid/service_bindings"
      }
    }
  ]
}`

    validBindings = `{
  "next_url": null,
  "resources": [
    {
      "entity": {
        "app_guid": "app-guid"
      }
    },
    {
      "entity": {
        "app_guid": "other"
      }
    }
  ]
}`

    validBindingsPage0 = `{
  "next_url": "/v2/user_provided_service_instances/guid/service_bindings",
  "resources": [
    {
      "entity": {
        "app_guid": "some-app-guid"
      }
    }
  ]
}`

    emptyBindings = `{
      "resources": []
    }`
)
