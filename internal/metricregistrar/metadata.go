package metricregistrar

import "code.cloudfoundry.org/cli/plugin"

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:     "metric-registrar",
		Version:  p.vt,
		Commands: []plugin.Command{},
	}
}
