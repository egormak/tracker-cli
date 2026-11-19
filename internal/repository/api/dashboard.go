package api

import (
	"encoding/json"
	"fmt"
	"tracker_cli/internal/domain/entity"
)

// GetDashboardState retrieves today's tasks and day totals from the server.
func GetDashboardState() (entity.DashboardState, error) {
	body, err := sendRequest("GET", "/api/v1/dashboard/state", nil)
	if err != nil {
		return entity.DashboardState{}, fmt.Errorf("get dashboard state: %w", err)
	}
	defer body.Close()

	var state entity.DashboardState
	if err := json.NewDecoder(body).Decode(&state); err != nil {
		return entity.DashboardState{}, fmt.Errorf("decode dashboard state: %w", err)
	}
	if state.TodayTasks == nil {
		state.TodayTasks = []entity.TaskList{}
	}

	return state, nil
}
