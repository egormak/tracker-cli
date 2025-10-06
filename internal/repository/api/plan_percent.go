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
