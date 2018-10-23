package command_test

import (
    "strings"

    "github.com/pivotal-cf/metric-registrar-cli/command"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"

    "code.cloudfoundry.org/cli/plugin/models"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/pkg/errors"
)

var _ = Describe("List", func() {
    It("displays registered log formats", func() {
        registrationFetcher := newMockRegistrationFetcher()
        registrationFetcher.registrations = map[string][]registrations.Registration{
            "app-guid": {
                {
                    Name:             "service1",
                    Type:             "structured-format",
                    Config:           "json",
                    NumberOfBindings: 2,
                },
                {
                    Name:             "service2",
                    Type:             "structured-format",
                    Config:           "dogstatsd",
                    NumberOfBindings: 2,
                }},
        }
        writer := newSpyWriter()
        cliConn := newMockCliConnection()
        cliConn.getAppsResult = []plugin_models.GetAppsModel{
            {Name: "app-name", Guid: "app-guid"},
        }

        err := command.ListRegisteredLogFormats(writer, registrationFetcher, cliConn)
        Expect(err).ToNot(HaveOccurred())

        Expect(writer.lines()).To(Equal([]string{
            "App       Format",
            "app-name  json",
            "app-name  dogstatsd",
            "",
        }))
    })

    It("returns an error if the fetcher fails", func() {
        registrationFetcher := newMockRegistrationFetcher()
        registrationFetcher.fetchError = errors.New("expected")

        writer := newSpyWriter()
        cliConn := newMockCliConnection()
        cliConn.getAppsResult = []plugin_models.GetAppsModel{
            {Name: "app-name", Guid: "app-guid"},
        }

        err := command.ListRegisteredLogFormats(writer, registrationFetcher, cliConn)
        Expect(err).To(HaveOccurred())
    })

    It("returns an error if the writer fails", func() {
        registrationFetcher := newMockRegistrationFetcher()
        registrationFetcher.registrations = map[string][]registrations.Registration{
            "app-guid": {{
                Name:             "service1",
                Type:             "structured-format",
                Config:           "json",
                NumberOfBindings: 2,
            }},
        }
        writer := newSpyWriter()
        writer.writeErr = errors.New("expected")
        cliConn := newMockCliConnection()
        cliConn.getAppsResult = []plugin_models.GetAppsModel{
            {Name: "app-name", Guid: "app-guid"},
        }

        err := command.ListRegisteredLogFormats(writer, registrationFetcher, cliConn)
        Expect(err).To(HaveOccurred())
    })

    It("returns an error if the app lister fails", func() {
        registrationFetcher := newMockRegistrationFetcher()
        registrationFetcher.registrations = map[string][]registrations.Registration{
            "app-guid": {{
                Name:             "service1",
                Type:             "structured-format",
                Config:           "json",
                NumberOfBindings: 2,
            }},
        }
        writer := newSpyWriter()
        cliConn := newMockCliConnection()
        cliConn.getAppsError = errors.New("expected")

        err := command.ListRegisteredLogFormats(writer, registrationFetcher, cliConn)
        Expect(err).To(HaveOccurred())
    })
})

type spyWriter struct {
    bytes    []byte
    writeErr error
}

func newSpyWriter() *spyWriter {
    return &spyWriter{}
}

func (s *spyWriter) Write(p []byte) (n int, err error) {
    s.bytes = append(s.bytes, p...)
    return len(p), s.writeErr
}

func (s *spyWriter) lines() []string {
    return strings.Split(string(s.bytes), "\n")
}
