package entity

// DashboardState holds the fields of GET /api/v1/dashboard/state the cli reads.
type DashboardState struct {
	TodayTasks    []TaskList `json:"today_tasks"`
	TotalPlanned  int        `json:"total_planned"`
	TotalDone     int        `json:"total_done"`
	CompletionPct int        `json:"completion_pct"`
}

// Task returns the first of today's tasks with the exact name.
func (s DashboardState) Task(name string) (TaskList, bool) {
	for _, t := range s.TodayTasks {
		if t.Name == name {
			return t, true
		}
	}
	return TaskList{}, false
}
