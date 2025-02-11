package api

import (
	"fmt"
	"io"
	"net/http"
	"time"
	"tracker_cli/config"
)

var timeout = time.Duration(15 * time.Second)

var client = http.Client{
	Timeout: timeout,
}

func sendRequest(method, path string, body io.Reader) (io.ReadCloser, error) {

	url := fmt.Sprintf("%s%s", config.TrackerDomain, path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error: status code %d", resp.StatusCode)
	}

	return resp.Body, nil

}
