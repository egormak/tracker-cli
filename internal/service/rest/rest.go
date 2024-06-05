package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"tracker_cli/config"
)

func RestSpend(restTime int) {
	slog.Info("Set Duration Rest", "rest_time", restTime)
	// Set Value
	values := map[string]int{"rest_time": restTime}
	json_data, err := json.Marshal(values)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/rest-spend"), bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("request error", "error", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	// Request
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}
}

func RestAdd(restTime int) {

	// Set Value
	values := map[string]int{"rest_time": restTime}
	json_data, err := json.Marshal(values)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/rest-add"), bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("request error", "error", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	// Request
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}
}

func RestShow() {

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/rest-get"), nil)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	// Request
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}

	var restTime map[string]int
	err = json.NewDecoder(resp.Body).Decode(&restTime)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
	}

	fmt.Println("Rest Earn: ", float64(restTime["rest_time"])/100)
}
