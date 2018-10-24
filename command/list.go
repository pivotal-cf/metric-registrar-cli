package command

import (
    "fmt"
    "io"
    "strings"
    "text/tabwriter"

    "code.cloudfoundry.org/cli/plugin/models"
    "github.com/pivotal-cf/metric-registrar-cli/registrations"
)

type appLister interface {
    GetApps() ([]plugin_models.GetAppsModel, error)
}

func ListRegisteredLogFormats(writer io.Writer, fetcher registrationFetcher, lister appLister, appName string) error {
    regs, err := fetcher.FetchAll(structuredFormat)
    if err != nil {
        return err
    }

    apps, err := lister.GetApps()
    if err != nil {
        return err
    }

    return writeTable(writer, apps, regs, appName, "Format")
}

func ListRegisteredMetricsEndpoints(writer io.Writer, fetcher registrationFetcher, lister appLister, appName string) error {
    regs, err := fetcher.FetchAll(metricsEndpoint)
    if err != nil {
        return err
    }

    apps, err := lister.GetApps()
    if err != nil {
        return err
    }

    return writeTable(writer, apps, regs, appName, "Path")
}

func writeTable(writer io.Writer, apps []plugin_models.GetAppsModel, regs map[string][]registrations.Registration, appName, configName string) error {
    w := tabwriter.NewWriter(writer, 0, 8, 2, ' ', tabwriter.StripEscape)
    writeFields(w, "App", configName)

    for _, line := range lines(apps, regs, appName) {
        writeFields(w, line...)
    }

    return w.Flush()
}

func lines(apps []plugin_models.GetAppsModel, regs map[string][]registrations.Registration, appName string) [][]string {
    var lines [][]string

    for _, app := range apps {
        for _, reg := range regs[app.Guid] {
            if appName == "" || appName == app.Name {
                lines = append(lines, []string{app.Name, reg.Config})
            }
        }
    }

    return lines
}

func writeFields(w *tabwriter.Writer, fields ...string) error {
    _, err := fmt.Fprintln(w, strings.Join(fields, "\t"))
    return err
}
