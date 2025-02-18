package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
)

func StatisticTaskGet(taskName string) int {

	var result entity.TaskTimeDurationResponse

	responceBody, err := sendRequest("GET", fmt.Sprintf("%s?task_name=%s", "/api/v1/record/task-day", taskName), nil)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result.TaskTime

}
