package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
)

func GetTaskParams(taskName string) entity.TaskParams {

	var result entity.TaskParams

	responceBody, err := sendRequest("GET", fmt.Sprintf("%s?task_name=%s", "/api/v1/task/params", taskName), nil)

	if err != nil {
		slog.Error("request error: get task params", "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result

}

func AddTaskRecord(taskName string, timeDone int) entity.Answer {

	taskRecord := entity.TaskRecorcRequest{TaskName: taskName, TimeDone: timeDone}
	var result entity.Answer

	json_data, err := json.Marshal(&taskRecord)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
		os.Exit(1)
	}

	responceBody, err := sendRequest("POST", "/api/v1/record", bytes.NewBuffer(json_data))

	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result

}

func GetTaskRecords() map[string]map[string]int {

	result := make(map[string]map[string]int)
	responceBody, err := sendRequest("GET", "/api/v1/records", nil)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	defer responceBody.Close()

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result

}
