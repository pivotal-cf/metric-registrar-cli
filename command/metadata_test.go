package command_test

import (
	"code.cloudfoundry.org/cli/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal-cf/metric-registrar-cli/command"
)

var _ = Describe("Metadata", func() {
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
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("unregister-log-format")}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("unregister-metrics-endpoint")}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("registered-log-formats")}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("registered-metrics-endpoints")}),
			))
		})
	})
})
