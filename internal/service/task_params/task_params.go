package task_params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"tracker_cli/config"
	"tracker_cli/internal/domain/entity"
)

type TaskParams = entity.TaskParams

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func SetTaskParams(taskName string, timeDur int, priority int) {
	if taskName == "" {
		slog.Error("Task Name is not Set")
		return
	}
	if timeDur == 0 {
		slog.Error("Time is not Set")
	}
	if priority == 0 {
		slog.Error("Priority is not Set")
	}

	body := entity.TaskParams{
		Name:     taskName,
		Time:     timeDur,
		Priority: priority,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
		return
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/task/params"), bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("request error", "error", err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
		return
	}
	if resp == nil {
		slog.Error("nil response received")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("request error", "status code", resp.StatusCode)
	}
}

func GetTaskParams(taskName string) TaskParams {
	requestURL := fmt.Sprintf("%s%s?task_name=%s", config.TrackerDomain, "/api/v1/task/params", url.QueryEscape(taskName))
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		slog.Error("error in request", "error", err)
		return TaskParams{}
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		slog.Error("error in request", "error", err)
		return TaskParams{}
	}
	if resp == nil {
		slog.Error("nil response received")
		return TaskParams{}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.Debug("task parameters not found", "task", taskName, "status_code", resp.StatusCode)
		return TaskParams{}
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("request error", "status code", resp.StatusCode)
		return TaskParams{}
	}

	var taskRole TaskParams
	err = json.NewDecoder(resp.Body).Decode(&taskRole)
	if err != nil {
		slog.Error("failed to decode response", "error", err)
		return TaskParams{}
	}

	return taskRole
}
