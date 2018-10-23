package command

import (
    "fmt"
    "io"
    "strings"
    "text/tabwriter"

    "code.cloudfoundry.org/cli/plugin/models"
)

type appLister interface {
    GetApps() ([]plugin_models.GetAppsModel, error)
}

func ListRegisteredLogFormats(writer io.Writer, fetcher registrationFetcher, lister appLister) error {
    registrations, err := fetcher.FetchAll(structuredFormat)
    if err != nil {
       return err
    }

    apps, err := lister.GetApps()
    if err != nil {
        return err
    }

    w := tabwriter.NewWriter(writer, 0, 8, 2, ' ', tabwriter.StripEscape)
    writeFields(w, "App", "Format")

    for _, app := range apps {
        for _, reg := range registrations[app.Guid] {
            writeFields(w, app.Name, reg.Config)
        }
    }

    return w.Flush()
}

func writeFields(w *tabwriter.Writer, fields ...string) error {
    _, err := fmt.Fprintln(w, strings.Join(fields, "\t"))
    return err
}
