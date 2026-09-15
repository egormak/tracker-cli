package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"tracker_cli/internal/domain/entity"
)

func StatisticTaskGet(taskName string) int {

	var result entity.TaskTimeDurationResponse

	responceBody, err := sendRequest("GET", fmt.Sprintf("%s?task_name=%s", "/api/v1/record/task-day", taskName), nil)
	if err != nil {
		slog.Error("request error", "error", err)
		return 0
	}
	defer responceBody.Close()

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response", "error", err)
		return 0
	}

	return result.TaskTime

}

func GetScheduledTimeToday() (int, error) {
	responseBody, err := sendRequest("GET", "/api/v1/manage/timer/global", nil)
	if err != nil {
		return 0, fmt.Errorf("get scheduled time today: %w", err)
	}
	defer responseBody.Close()

	var timerGlobal map[string]int
	if err := json.NewDecoder(responseBody).Decode(&timerGlobal); err != nil {
		return 0, fmt.Errorf("decode scheduled time response: %w", err)
	}

	return timerGlobal["timer_global"], nil
}

func GetCompletionTimeDoneToday() (int, error) {
	responseBody, err := sendRequest("GET", "/api/v1/roles/records/today", nil)
	if err != nil {
		return 0, fmt.Errorf("get completion time today: %w", err)
	}
	defer responseBody.Close()

	var timerDone map[string]int
	if err := json.NewDecoder(responseBody).Decode(&timerDone); err != nil {
		return 0, fmt.Errorf("decode completion time response: %w", err)
	}

	return timerDone["time_done"], nil
}

