package procent

import (
	"fmt"

	"tracker_cli/internal/repository/api"
)

// ChangeGroupPlanPercent notifies the backend that a task percent group was consumed.
func ChangeGroupPlanPercent() (string, error) {
	message, err := api.TriggerPlanPercentChange()
	if err != nil {
		return "", fmt.Errorf("notify plan-percent change: %w", err)
	}
	return message, nil
}
