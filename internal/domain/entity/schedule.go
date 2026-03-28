package entity

type RolloverTask struct {
	TaskName      string `json:"task_name"`
	Role          string `json:"role"`
	Priority      int    `json:"priority"`
	RemainingTime int    `json:"remaining_time"`
	SourceDay     string `json:"source_day"`
	Percent       int    `json:"percent"`
}

type RolloverResponse struct {
	Status string `json:"status"`
	Data   struct {
		Day           string         `json:"day"`
		RolloverTasks []RolloverTask `json:"rollover_tasks"`
		Count         int            `json:"count"`
	} `json:"data"`
}
