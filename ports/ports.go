package ports

import (
	"encoding/json"
	"fmt"
	"strings"
)

type cliConn interface {
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
}

type Response struct {
	Entity Entity `json:"entity"`
}

type Entity struct {
	Ports []int `json:"ports"`
}

func GetPortsForApp(cliConn cliConn, guid string) ([]int, error) {
	appsEndpoint := fmt.Sprintf("/v2/apps/%s", guid)
	output, err := cliConn.CliCommandWithoutTerminalOutput("curl", appsEndpoint)
	if err != nil {
		return []int{}, err
	}
	joined := strings.Join(output, "")
	response := Response{}
	err = json.Unmarshal([]byte(joined), &response)
	return response.Entity.Ports, err
}

func SetPortsForApp(cliConn cliConn, guid string, ports []int) error {
	appsEndpoint := fmt.Sprintf("/v2/apps/%s", guid)

	newPortsEntity := Entity{Ports: ports}
	portsBody, err := json.Marshal(newPortsEntity)
	if err != nil {
		return err
	}

	wrappedPortsBody := fmt.Sprintf("'%s'", string(portsBody))
	_, err = cliConn.CliCommandWithoutTerminalOutput("curl", appsEndpoint, "-X", "PUT", "-d", wrappedPortsBody)
	return err
}
