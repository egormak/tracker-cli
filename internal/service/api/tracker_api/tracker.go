package tracker_api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"tracker_cli/config"
)

type APIClient struct {
	Method string
	Url    string
	Body   []byte
}

func NewClient(method string, url string, body []byte) *APIClient {
	return &APIClient{
		Method: method,
		Url:    url,
		Body:   body,
	}
}

func (c *APIClient) Request() ([]byte, error) {
	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/records/clean"), nil)
	resp, err := client.Do(request)

	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status", resp.StatusCode)
		os.Exit(1)
	}
	return nil, nil
}
