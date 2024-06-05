package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"tracker_cli/config"
)

func AddTaskRecord(taskName string, timeDone int) {

	type taskRecord struct {
		TaskName string `json:"task_name"`
		TimeDone int    `json:"time_done"`
	}

	// Check Value
	if taskName == "" {
		slog.Error("Task name is not Set")
		os.Exit(1)
	}
	if timeDone == 0 {
		slog.Info("Task Duration is Zero")
		return
	}

	// Set Value
	values := taskRecord{TaskName: taskName, TimeDone: timeDone}
	json_data, err := json.Marshal(&values)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
		os.Exit(1)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/record"), bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
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
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}
}

func GetTaskByProcentPlan() TaskManager {

	var result struct {
		TaskName string `json:"task_name"`
		Percent  int    `json:"percent"`
	}

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/task/plan-percent"), nil)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
		os.Exit(1)
	}

	return TasksNew(result.TaskName, 0, result.Percent)
}

func GetTaskDay(percent int) TaskManager {

	var result struct {
		TaskName string `json:"task_name"`
	}

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s?percent=%d", config.TrackerDomain, "/api/v1/task/percent", percent), nil)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
		os.Exit(1)
	}

	return TasksNew(result.TaskName, 0, percent)
}
