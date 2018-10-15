package registrations

import (
    "code.cloudfoundry.org/cli/plugin/models"
    "encoding/json"
    "fmt"
    "strings"
)

type cliConn interface {
    CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
    GetCurrentSpace() (plugin_models.Space, error)
}

type servicesResponse struct {
    Entity serviceEntity `json:"entity"`
}

type serviceEntity struct {
    Name               string `json:"name"`
    DrainUrl           string `json:"syslog_drain_url"`
    ServiceBindingsUrl string `json:"service_bindings_url"`
}

type bindingsResponse struct {
    Entity bindingEntity `json:"entity"`
}

type bindingEntity struct {
    AppGuid string `json:"app_guid"`
}

type Registration struct {
    Name             string
    Type             string
    Config           string
    NumberOfBindings int
}

type Fetcher struct {
    cliConn cliConn
}

func NewFetcher(conn cliConn) *Fetcher {
    return &Fetcher{cliConn: conn}
}

func (f *Fetcher) Fetch(appGuid, registrationType string) ([]Registration, error) {
    services, err := f.getServices(appGuid)
    if err != nil {
        return nil, err
    }

    var result []Registration
    for _, s := range services {
        r, isRegistration := registration(s.Entity)
        if isRegistration && r.Type == registrationType {
            bindings, err := f.serviceBindings(s.Entity.ServiceBindingsUrl)
            if err != nil {
                return nil, err
            }
            if !f.isBound(appGuid, bindings) {
                continue
            }

            r.NumberOfBindings = len(bindings)
            result = append(result, r)
        }
    }

    return result, nil
}

func (f *Fetcher) getServices(appGuid string) (services []servicesResponse, err error) {
    space, err := f.cliConn.GetCurrentSpace()
    if err != nil {
        return services, err
    }

    path := fmt.Sprintf("/v2/user_provided_service_instances?q=space_guid:%s", space.Guid)
    err = f.getPagedResource(path, func(messages json.RawMessage) error {
        var page []servicesResponse

        err := json.Unmarshal(messages, &page)
        if err != nil {
            return err
        }
        services = append(services, page...)
        return nil
    })
    return services, err
}

func registration(e serviceEntity) (Registration, bool) {
    drainUrlComponents := strings.Split(e.DrainUrl, "://")
    if len(drainUrlComponents) != 2 {
        return Registration{}, false
    }

    return Registration{
        Name:   e.Name,
        Type:   drainUrlComponents[0],
        Config: drainUrlComponents[1],
    }, true
}

func (f *Fetcher) isBound(appGuid string, bindings []bindingsResponse) bool {
    for _, b := range bindings {
        if b.Entity.AppGuid == appGuid {
            return true
        }
    }
    return false
}

func (f *Fetcher) serviceBindings(serviceBindingsUrl string) (bindings []bindingsResponse, err error) {
    err = f.getPagedResource(serviceBindingsUrl, func(messages json.RawMessage) error {
        var page []bindingsResponse

        err := json.Unmarshal(messages, &page)
        if err != nil {
            return err
        }
        bindings = append(bindings, page...)
        return nil
    })
    return bindings, err
}

type accumulator func(json.RawMessage) error

type paginatedResp struct {
    Resources json.RawMessage `json:"resources"`
    NextUrl   *string         `json:"next_url"`
}

func (f *Fetcher) getPagedResource(path string, a accumulator) error {
    var err error
    for path != "" {
        path, err = f.getPage(path, a)
        if err != nil {
            return err
        }
    }

    return nil
}

func (f *Fetcher) getPage(path string, a accumulator) (string, error) {
    resp, err := f.cliConn.CliCommandWithoutTerminalOutput("curl", path)
    if err != nil {
        return "", err
    }

    var page paginatedResp
    err = json.Unmarshal([]byte(strings.Join(resp, "")), &page)
    if err != nil {
        return "", err
    }

    err = a(page.Resources)
    if err != nil {
        return "", err
    }

    if page.NextUrl != nil {
        return *page.NextUrl, nil
    }

    return "", nil
}
