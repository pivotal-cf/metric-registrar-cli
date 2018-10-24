package command

import (
    "fmt"
    "os"
    "strings"

    "code.cloudfoundry.org/cli/plugin"
)

const (
    registerLogFormatCommand         = "register-log-format"
    registerMetricsEndpointCommand   = "register-metrics-endpoint"
    unregisterLogFormatCommand       = "unregister-log-format"
    unregisterMetricsEndpointCommand = "unregister-metrics-endpoint"
    listLogFormatsCommand            = "registered-log-formats"
)

type Command struct {
    name string

    HelpText  string
    Arguments []string
    Options   map[string]Option

    Flags interface{}
    Run   func(fetcher registrationFetcher, conn plugin.CliConnection) error
}

func (c Command) Usage() string {
    argsAndOptions := append([]string{}, c.Arguments...)
    for flag, opt := range c.Options {
        argsAndOptions = append(argsAndOptions, fmt.Sprintf("[%s %s]", flag, opt.Name))
    }

    return fmt.Sprintf("cf %s %s",
        c.name,
        strings.Join(argsAndOptions, " "),
    )
}

type Option struct {
    Name        string
    Description string
}

var registerLogFormatFlags = &struct {
    Args struct {
        AppName string `positional-arg-name:"APP_NAME"`
        Format  string `positional-arg-name:"FORMAT"`
    } `positional-args:"APP_NAME FORMAT" required:"2"`
}{}
var registerMetricsEndpointFlags = &struct {
    Args struct {
        AppName string `positional-arg-name:"APP_NAME"`
        Path    string `positional-arg-name:"PATH"`
    } `positional-args:"APP_NAME PATH" required:"2"`
}{}
var unregisterMetricsEndpointFlags = &struct {
    Path string                                                    `short:"p" long:"path"`
    Args struct{ AppName string `positional-arg-name:"APP_NAME"` } `positional-args:"APP_NAME" required:"1"`
}{}
var unregisterLogFormatFlags = &struct {
    Format string                                                    `short:"f" long:"format"`
    Args   struct{ AppName string `positional-arg-name:"APP_NAME"` } `positional-args:"APP_NAME" required:"1"`
}{}

var Registry = map[string]Command{
    registerLogFormatCommand: {
        name:      registerLogFormatCommand,
        HelpText:  "Register bound applications so that structured logs of the given format can be parsed",
        Arguments: []string{"APP_NAME", "<json|DogStatsD>"},
        Flags:     registerLogFormatFlags,
        Run: func(_ registrationFetcher, conn plugin.CliConnection) error {
            return RegisterLogFormat(
                conn,
                registerLogFormatFlags.Args.AppName,
                registerLogFormatFlags.Args.Format,
            )
        },
    },
    registerMetricsEndpointCommand: {
        HelpText:  "Register a metrics endpoint which will be scraped at the interval defined at deploy",
        Arguments: []string{"APP_NAME", "PATH"},
        Flags:     registerMetricsEndpointFlags,
        Run: func(_ registrationFetcher, conn plugin.CliConnection) error {
            return RegisterMetricsEndpoint(
                conn,
                registerMetricsEndpointFlags.Args.AppName,
                registerMetricsEndpointFlags.Args.Path,
            )
        },
    },
    unregisterLogFormatCommand: {
        HelpText: "Unregister log formats",
        Options: map[string]Option{
            "-f": {
                Name:        "FORMAT",
                Description: "unregister only the specified log format",
            },
        },
        Flags: unregisterLogFormatFlags,
        Run: func(fetcher registrationFetcher, conn plugin.CliConnection) error {
            return UnregisterLogFormat(
                fetcher,
                conn,
                unregisterLogFormatFlags.Args.AppName,
                unregisterLogFormatFlags.Format,
            )
        },
    },
    unregisterMetricsEndpointCommand: {
        HelpText: "Unregister metrics endpoints",
        Options: map[string]Option{
            "-p": {
                Name:        "PATH",
                Description: "unregister only the specified path",
            },
        },
        Flags: unregisterMetricsEndpointFlags,
        Run: func(fetcher registrationFetcher, conn plugin.CliConnection) error {
            return UnregisterMetricsEndpoint(
                fetcher,
                conn,
                unregisterMetricsEndpointFlags.Args.AppName,
                unregisterMetricsEndpointFlags.Path,
            )
        },
    },
    listLogFormatsCommand: {
        HelpText: "List log formats in space",
        Run: func(fetcher registrationFetcher, conn plugin.CliConnection) error {
            return ListRegisteredLogFormats(os.Stdout, fetcher, conn)
        },
    },
}
