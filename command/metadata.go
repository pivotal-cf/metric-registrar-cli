package command

import (
    "code.cloudfoundry.org/cli/plugin"
)

func (c MetricRegistrarCli) GetMetadata() plugin.PluginMetadata {
    var commands []plugin.Command
    for name, c := range Registry {
        commands = append(commands, plugin.Command{
            Name:     name,
            HelpText: c.HelpText,
            UsageDetails: plugin.Usage{
                Usage:   c.Usage(),
                Options: buildOptions(c),
            },
        })
    }

    return plugin.PluginMetadata{
        Name: pluginName,
        Version: plugin.VersionType{
            Major: c.Major,
            Minor: c.Minor,
            Build: c.Patch,
        },
        Commands: commands,
    }
}

func buildOptions(c Command) map[string]string {
    opts := map[string]string{}
    for flag, opt := range c.Options {
        opts[flag] = opt.Description
    }
    return opts
}
