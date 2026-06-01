package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"tracker_cli/internal/domain/entity"
)

type runningTaskResponse struct {
	Status string             `json:"status"`
	Data   entity.RunningTask `json:"data"`
}

type taskRecordResponse struct {
	Status string            `json:"status"`
	Data   entity.TaskRecord `json:"data"`
}

func StartRunningTask(name, role string, targetDuration int, sourceDay string) (entity.RunningTask, error) {
	payload := struct {
		TaskName       string `json:"task_name"`
		Role           string `json:"role"`
		TargetDuration int    `json:"target_duration"`
		SourceDay      string `json:"source_day"`
	}{
		TaskName:       name,
		Role:           role,
		TargetDuration: targetDuration,
		SourceDay:      sourceDay,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return entity.RunningTask{}, fmt.Errorf("marshal start request payload: %w", err)
	}

	responseBody, err := sendRequest("POST", "/api/v1/timer/run/start", bytes.NewBuffer(jsonData))
	if err != nil {
		return entity.RunningTask{}, fmt.Errorf("start running task: %w", err)
	}
	defer responseBody.Close()

	var resp runningTaskResponse
	if err := json.NewDecoder(responseBody).Decode(&resp); err != nil {
		return entity.RunningTask{}, fmt.Errorf("decode start response: %w", err)
	}

	return resp.Data, nil
}

func GetRunningTaskStatus() (entity.RunningTask, error) {
	responseBody, err := sendRequest("GET", "/api/v1/timer/run/status", nil)
	if err != nil {
		return entity.RunningTask{}, fmt.Errorf("get running task status: %w", err)
	}
	defer responseBody.Close()

	var resp runningTaskResponse
	if err := json.NewDecoder(responseBody).Decode(&resp); err != nil {
		return entity.RunningTask{}, fmt.Errorf("decode status response: %w", err)
	}

	return resp.Data, nil
}

func PauseRunningTask() (entity.RunningTask, error) {
	responseBody, err := sendRequest("POST", "/api/v1/timer/run/pause", nil)
	if err != nil {
		return entity.RunningTask{}, fmt.Errorf("pause running task: %w", err)
	}
	defer responseBody.Close()

	var resp runningTaskResponse
	if err := json.NewDecoder(responseBody).Decode(&resp); err != nil {
		return entity.RunningTask{}, fmt.Errorf("decode pause response: %w", err)
	}

	return resp.Data, nil
}

func ResumeRunningTask() (entity.RunningTask, error) {
	responseBody, err := sendRequest("POST", "/api/v1/timer/run/resume", nil)
	if err != nil {
		return entity.RunningTask{}, fmt.Errorf("resume running task: %w", err)
	}
	defer responseBody.Close()

	var resp runningTaskResponse
	if err := json.NewDecoder(responseBody).Decode(&resp); err != nil {
		return entity.RunningTask{}, fmt.Errorf("decode resume response: %w", err)
	}

	return resp.Data, nil
}

func StopRunningTask() (entity.TaskRecord, error) {
	responseBody, err := sendRequest("POST", "/api/v1/timer/run/stop", nil)
	if err != nil {
		return entity.TaskRecord{}, fmt.Errorf("stop running task: %w", err)
	}
	defer responseBody.Close()

	var resp taskRecordResponse
	if err := json.NewDecoder(responseBody).Decode(&resp); err != nil {
		return entity.TaskRecord{}, fmt.Errorf("decode stop response: %w", err)
	}

	return resp.Data, nil
}
