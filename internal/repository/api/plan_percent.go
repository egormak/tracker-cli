package api

import (
	"encoding/json"
	"fmt"
	"tracker_cli/internal/domain/entity"
)

func GetTaskByPercentPlan() (string, int, error) {
	var result entity.TaskPercent

	responseBody, err := sendRequest("GET", "/api/v1/task/plan-percent", nil)
	if err != nil {
		return "", 0, fmt.Errorf("request percent plan: %w", err)
	}
	defer responseBody.Close()

	if err := json.NewDecoder(responseBody).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("decode percent plan response: %w", err)
	}

	return result.Name, result.Percent, nil
}

// GetTaskByPercentPlanSchedule gets the next task based on plan percentage with schedule awareness.
// Returns task name, percent, time left (in minutes), source_day, and error if any.
func GetTaskByPercentPlanSchedule() (string, int, int, string, error) {
	var result entity.TaskPercent

	responseBody, err := sendRequest("GET", "/api/v1/task/plan/percent/schedule", nil)
	if err != nil {
		return "", 0, 0, "", fmt.Errorf("request schedule-aware percent plan: %w", err)
	}
	defer responseBody.Close()

	if err := json.NewDecoder(responseBody).Decode(&result); err != nil {
		return "", 0, 0, "", fmt.Errorf("decode schedule percent plan response: %w", err)
	}

	return result.Name, result.Percent, result.TimeLeft, result.SourceDay, nil
}
