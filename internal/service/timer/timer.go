package timer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
	"tracker_cli/config"
)

// TimeListSet sends a POST request to the specified URL with a JSON payload containing the timerCount value.
func TimeListSet(timerCount int) {
	// Create a map with the timerCount value.
	values := map[string]int{"count": timerCount}

	// Convert the map to JSON.
	jsonData, err := json.Marshal(values)
	if err != nil {
		slog.Error("error in json marshal", "error", err)
	}

	// Construct the URL for the POST request.
	url := fmt.Sprintf("%s%s", config.TrackerDomain, "/v1/timer/set")

	// Create a new POST request.
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("error in new request", "error", err)
	}

	// Set the Content-Type and Accept headers.
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// Create an HTTP client with a 15-second timeout.
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	// Send the request.
	slog.Info("Sending request", "URL", url)
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("error in request", "error", err)
	}

	defer resp.Body.Close()

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		slog.Error("request error", "status code", resp.StatusCode)
	}

	slog.Info("Request successful")
}

// Recheck Count Timers
func TimerRecheck() {

	slog.Info("Timer Recheck")

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/manage/timer/recheck"), nil)
	resp, err := client.Do(request)

	if err != nil {
		slog.Error("error in request", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}

	var Data map[string]string

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &Data); err != nil {
		log.Fatal(err)
	}

	slog.Info(Data["message"])

}

func SetGlobalTime(timeScheduler int) {

	var values map[string]int

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	values = map[string]int{"time_scheduler": timeScheduler}
	json_data, err := json.Marshal(&values)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/manage/timer/global"), bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal(fmt.Errorf("request error, status code: %d", resp.StatusCode))
	}
}

func TimeDurationGet() int {

	slog.Info("Get Time Duration")

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/timer/get"), nil)
	resp, err := client.Do(request)

	if err != nil {
		slog.Error("error in request", "error", err)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var Data map[string]int

	if err := json.Unmarshal(body, &Data); err != nil {
		slog.Error("can't unmarshal JSON", "error", err)
		os.Exit(1)
	}

	return Data["time_duration"]

}

func TimeDurationDel(timerCount int) {
	// Create a map with the timerCount value.
	values := map[string]int{"count": timerCount}

	// Convert the map to JSON.
	jsonData, err := json.Marshal(values)
	if err != nil {
		slog.Error("error in json marshal", "error", err)
		os.Exit(1)
	}

	// Construct the URL for the POST request.
	url := fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/timer/del")

	// Create a new POST request.
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("error in new request", "error", err)
		os.Exit(1)
	}

	// Set the Content-Type and Accept headers.
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// Create an HTTP client with a 15-second timeout.
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	// Send the request.
	slog.Info("Sending request", "URL", url)
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("error in request", "error", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}
	slog.Info("Request successful")
}
