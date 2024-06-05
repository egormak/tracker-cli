package task_params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"tracker_cli/config"
)

type TaskParams struct {
	Name     string
	Time     int
	Priority int
}

func SetTaskParams(taskName string, timeDur int, priority int) {

	var body struct {
		Name         string `json:"name"`
		TimeDuration int    `json:"time_duration"`
		Priority     int    `json:"priority"`
	}

	if taskName == "" {
		slog.Error("Task Name is not Set")
	}
	if timeDur == 0 {
		slog.Error("Time is not Set")
	}
	if priority == 0 {
		slog.Error("Priority is not Set")
	}

	body.Name = taskName
	body.TimeDuration = timeDur
	body.Priority = priority

	json_data, err := json.Marshal(&body)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/record/params"), bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("request error", "error", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}

}

func GetTaskParams(taskName string) TaskParams {

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s?task_name=%s", config.TrackerDomain, "/api/v1/record/params", taskName), nil)
	if err != nil {
		slog.Error("error in request", "error", err)
	}
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("error in request", "error", err)
	}

	if resp.StatusCode == 404 {
		slog.Info("Task Params not found", "status code", resp.StatusCode)
		return TaskParams{}
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}
	defer resp.Body.Close()

	var taskRole TaskParams
	err = json.NewDecoder(resp.Body).Decode(&taskRole)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
	}

	return taskRole
}
