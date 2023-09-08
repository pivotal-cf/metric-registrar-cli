/*
Package metricregistrar is an internal package that provides the
metric-registrar cf CLI plugin. It is not intended to be used directly outside
of this repository.
*/
package metricregistrar

import "code.cloudfoundry.org/cli/plugin"

type Plugin struct {
	vt plugin.VersionType
}

func NewPlugin(vt plugin.VersionType) *Plugin {
	return &Plugin{vt: vt}
}
