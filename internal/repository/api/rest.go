package api

import (
	"encoding/json"
	"fmt"
)

type restTimeResponse struct {
	RestTime int `json:"rest_time"`
}

// GetRestTime returns the current rest balance in raw API units.
func GetRestTime() (int, error) {
	body, err := sendRequest("GET", "/api/v1/rest-get", nil)
	if err != nil {
		return 0, fmt.Errorf("request rest time: %w", err)
	}
	defer body.Close()

	var resp restTimeResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return 0, fmt.Errorf("decode rest time response: %w", err)
	}

	return resp.RestTime, nil
}
