package procent

import (
	"fmt"
	"tracker_cli/internal/repository/api"
)

// ChangeGroupPlanPercent notifies the backend that a task percent group was consumed.
func ChangeGroupPlanPercent() error {
	if err := api.TriggerPlanPercentChange(); err != nil {
		return fmt.Errorf("notify plan-percent change: %w", err)
	}
	return nil
}
