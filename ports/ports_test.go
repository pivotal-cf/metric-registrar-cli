package ports_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pivotal-cf/metric-registrar-cli/ports"
)

var _ = Describe("Ports", func() {
	It("gets the exposed ports of the application", func() {
		p1 := []int{1234, 2345}
		cliConn := newMockCliConnection(p1)
		p2, err := ports.GetPortsForApp(cliConn, "app-guid")
		Expect(err).ToNot(HaveOccurred())
		Expect(p2).To(Equal(p1))
		expectToReceiveGetCurlForAppAndPort(cliConn.cliCommandsCalled, "app-guid")
	})

	It("sets the exposed ports of the application", func() {
		p1 := []int{1234, 5678}
		cliConn := newMockCliConnection([]int{})
		err := ports.SetPortsForApp(cliConn, "app-guid", p1)
		Expect(err).ToNot(HaveOccurred())
		expectToReceivePutCurlForAppAndPort(cliConn.cliCommandsCalled, "app-guid", p1)
	})
})

type mockCliConnection struct {
	curlResponses     []string
	curlErrors        map[string]error
	cliCommandsCalled chan []string
}

func newMockCliConnection(ports []int) *mockCliConnection {
	stringResponse := fmt.Sprintf(`{"entity": {"ports": %v}}`, transformIntSlice(ports))
	response := strings.Split(stringResponse, "\n")

	return &mockCliConnection{
		cliCommandsCalled: make(chan []string, 10),
		curlResponses:     response,
		curlErrors:        map[string]error{},
	}
}
func (c *mockCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	c.cliCommandsCalled <- args
	Expect(args[0]).To(Equal("curl"))
	return c.curlResponses, nil
}

func expectToReceivePutCurlForAppAndPort(called chan []string, appGuid string, ports []int) {
	Eventually(called).Should(Receive(matchCurl(
		fmt.Sprintf("/v2/apps/%s", appGuid),
		"-X",
		"PUT",
		"-d",
		fmt.Sprintf("'{\"ports\":%s}'", transformIntSlice(ports)),
	)))
}

func expectToReceiveGetCurlForAppAndPort(called chan []string, appGuid string) {
	Eventually(called).Should(Receive(matchCurl(fmt.Sprintf("/v2/apps/%s", appGuid))))
}

func matchCurl(args ...string) types.GomegaMatcher {
	return Equal(append([]string{"curl"}, args...))
}

func transformIntSlice(i []int) string {
	return strings.Replace(fmt.Sprintf("%v", i), " ", ",", -1)
}
