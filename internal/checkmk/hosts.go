package checkmk

import "net/http"

const (
	apiCreateHostEndpoint = "domain-types/host_config/collections/all"
	apiGetHostEndpoint    = "objects/host_config/"
)

type Labels map[string]string

type HostAttributes struct {
	Parents          []string `json:"parents"`
	TagAddressFamily string   `json:"tag_address_family"`
	TagAgent         string   `json:"tag_agent"`
	TagPiggyback     string   `json:"tag_piggyback"`
	Labels           Labels   `json:"labels"`
}

type Host struct {
	Hostname       string `json:"host_name"`
	Folder         string `json:"folder"`
	HostAttributes `json:"attributes"`
}

func newDockerHost(hostname string, folder string, parent string, labels Labels) Host {
	labels["cmk-piggyback-docker-resolver"] = "true"
	return Host{hostname, folder, HostAttributes{
		Parents:          []string{parent},
		TagAddressFamily: "no-ip",
		TagAgent:         "no-agent",
		TagPiggyback:     "piggyback",
		Labels:           labels,
	}}
}

func (c *api) CreateHost(hostname string, folder string, parent string, labels Labels) error {
	host := newDockerHost(hostname, folder, parent, labels)
	status, err := c.request(http.MethodPost, apiCreateHostEndpoint, host)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return ErrCheckmkApiError
	}

	return nil
}

func (c *api) GetHost(hostname string) (bool, error) {
	status, err := c.request(http.MethodGet, apiGetHostEndpoint+hostname, nil)
	if err != nil {
		return false, err
	}

	switch status {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, ErrCheckmkApiError
	}
}
