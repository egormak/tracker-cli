package api

import (
	"encoding/json"
	"fmt"
	"tracker_cli/internal/domain/entity"
)

// GetDebt fetches this week's debt: the previous days' unfinished schedule and the week strip.
func GetDebt() (entity.DebtSummary, error) {
	body, err := sendRequest("GET", "/api/v1/debt", nil)
	if err != nil {
		return entity.DebtSummary{}, err
	}
	defer body.Close()

	var resp struct {
		Status string             `json:"status"`
		Data   entity.DebtSummary `json:"data"`
	}
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return entity.DebtSummary{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return resp.Data, nil
}
