package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"tracker_cli/config"
	"tracker_cli/internal/domain/entity"
)

func GetRolloverTasks() ([]entity.RolloverTask, error) {
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/schedule/active/rollover"), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error in request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var response entity.RolloverResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("can't unmarshal JSON: %w", err)
	}

	return response.Data.RolloverTasks, nil
}

func UpdateScheduleTaskTime(taskName string, minutes int, mode string, day string) error {
	payload := map[string]interface{}{
		"task_name": taskName,
		"minutes":   minutes,
		"mode":      mode,
	}
	if day != "" {
		payload["day"] = day
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	body, err := sendRequest("PATCH", "/api/v1/schedule/active/task-time", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("update schedule task time: %w", err)
	}
	defer body.Close()
	return nil
}

