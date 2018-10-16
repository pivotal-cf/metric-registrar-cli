package command

import "code.cloudfoundry.org/cli/plugin"

func (c MetricRegistrarCli) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name: pluginName,
        Version: plugin.VersionType{
            Major: c.Major,
            Minor: c.Minor,
            Build: c.Patch,
        },
        Commands: []plugin.Command{
            {
                Name:     registerLogFormatCommand,
                HelpText: "Register bound applications so that structured logs of the given format can be parsed",
                UsageDetails: plugin.Usage{
                    Usage: registerLogFormatUsage,
                },
            },
            {
                Name:     registerMetricsEndpointCommand,
                HelpText: "Register a metrics endpoint which will be scraped at the interval defined at deploy",
                UsageDetails: plugin.Usage{
                    Usage: registerMetricsEndpointUsage,
                },
            },
            {
                Name:     unregisterLogFormatCommand,
                HelpText: "Unregister log formats",
                UsageDetails: plugin.Usage{
                    Usage: unregisterLogFormatUsage,
                    Options: map[string]string{
                        "-f": "unregister only the specified log format",
                    },
                },
            },
            {
                Name:     unregisterMetricsEndpointCommand,
                HelpText: "Unregister metrics endpoints",
                UsageDetails: plugin.Usage{
                    Usage: unregisterMetricsEndpointUsage,
                    Options: map[string]string{
                        "-p": "unregister only the specified path",
                    },
                },
            },
        },
    }
}
