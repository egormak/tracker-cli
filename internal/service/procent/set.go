package procent

import (
	"errors"
	"fmt"
	"log/slog"
	"tracker_cli/internal/repository/api"
)

// SetRolePercents updates the percent distribution for the given role.
func SetRolePercents(role string, percents []int) error {
	if role == "" {
		return errors.New("role name is required")
	}
	if len(percents) == 0 {
		return errors.New("provide at least one percent value")
	}

	for i, p := range percents {
		if p < 0 {
			return fmt.Errorf("percent[%d] must be non-negative", i)
		}
	}

	if err := api.UpdateRoleProcents(role, percents); err != nil {
		return err
	}

	slog.Info("updated role percents", "role", role, "percents", percents)
	return nil
}
