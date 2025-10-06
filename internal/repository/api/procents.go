package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type roleProcentsPayload struct {
	RoleName string `json:"role_name"`
	Procents []int  `json:"procents"`
}

// UpdateRoleProcents sends the percent distribution for the given role to the backend.
func UpdateRoleProcents(role string, percents []int) error {
	payload := roleProcentsPayload{RoleName: role, Procents: percents}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal role percents: %w", err)
	}

	respBody, err := sendRequest("POST", "/api/v1/manage/procents", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post role percents: %w", err)
	}
	defer respBody.Close()

	if _, err := io.Copy(io.Discard, respBody); err != nil {
		return fmt.Errorf("drain response body: %w", err)
	}

	return nil
}

// TriggerPlanPercentChange requests a refresh of the percent plan on the backend.
func TriggerPlanPercentChange() error {
	respBody, err := sendRequest("GET", "/api/v1/task/plan-percent/change", nil)
	if err != nil {
		return fmt.Errorf("trigger plan percent change: %w", err)
	}
	defer respBody.Close()

	if _, err := io.Copy(io.Discard, respBody); err != nil {
		return fmt.Errorf("drain change response: %w", err)
	}

	return nil
}
