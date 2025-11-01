package checkmk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gringolito/cmk-piggyback-docker-resolver/pkg/logx"
)

var (
	ErrCheckmkApiError = errors.New("checkmk API returned an error")
)

const (
	fmtCheckmkAPIURL = `%s://%s/%s/check_mk/api/1.0/`
)

func NewAPIClient(protocol string, host string, site string, username string, secret string, log *logx.Logger) Client {
	apiURL := fmt.Sprintf(fmtCheckmkAPIURL, protocol, host, site)
	c := &http.Client{}
	return &api{c, apiURL, username, secret, log}
}

type api struct {
	c        *http.Client
	apiURL   string
	username string
	secret   string
	log      *logx.Logger
}

func (c *api) request(method string, endpoint string, data any) (int, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(method, c.apiURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.SetBasicAuth(c.username, c.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.c.Do(req)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
