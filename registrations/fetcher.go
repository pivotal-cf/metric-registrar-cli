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
    Resources []struct {
        Entity serviceEntity `json:"entity"`
    } `json:"resources"`
}

type serviceEntity struct {
    Name               string `json:"name"`
    DrainUrl           string `json:"syslog_drain_url"`
    ServiceBindingsUrl string `json:"service_bindings_url"`
}

type bindingsResponse struct {
    Resources []struct {
        Entity bindingEntity `json:"entity"`
    } `json:"resources"`
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
    for _, s := range services.Resources {
        r, isRegistration := registration(s.Entity)
        if isRegistration && r.Type == registrationType {
            bindings, err := f.serviceBindings(s.Entity.ServiceBindingsUrl)
            if err != nil {
                return nil, err
            }
            if !f.isBound(appGuid, bindings) {
                continue
            }

            r.NumberOfBindings = len(bindings.Resources)
            result = append(result, r)
        }
    }

    return result, nil
}

func (f *Fetcher) getServices(appGuid string) (services servicesResponse, err error) {
    space, err := f.cliConn.GetCurrentSpace()
    if err != nil {
        return services, err
    }

    path := fmt.Sprintf("/v2/user_provided_service_instances?q=space_guid:%s", space.Guid)
    resp, err := f.cliConn.CliCommandWithoutTerminalOutput("curl", path)
    if err != nil {
        return services, err
    }

    err = json.Unmarshal([]byte(strings.Join(resp, "")), &services)
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

func (f *Fetcher) isBound(appGuid string, bindings bindingsResponse) bool {
    for _, b := range bindings.Resources {
        if b.Entity.AppGuid == appGuid {
            return true
        }
    }
    return false
}

func (f *Fetcher) serviceBindings(serviceBindingsUrl string) (bindings bindingsResponse, err error) {
    resp, err := f.cliConn.CliCommandWithoutTerminalOutput("curl", serviceBindingsUrl)
    if err != nil {
        return bindings, err
    }

    err = json.Unmarshal([]byte(strings.Join(resp, "")), &bindings)
    if err != nil {
        return bindings, err
    }

    return bindings, nil

}