package entity

type EveningFocusCandidate struct {
	TaskName  string `json:"task_name"`
	Role      string `json:"role"`
	WeeklyGap int    `json:"weekly_gap"`
	Priority  int    `json:"priority"`
	IsStrict  bool   `json:"is_strict"`
}

type EveningFocusResponse struct {
	CurrentTask EveningFocusCandidate   `json:"current_task"`
	Candidates  []EveningFocusCandidate `json:"candidates"`
	SprintTime  int                     `json:"sprint_time"`
	RestPool    int                     `json:"rest_pool"`
}

type EveningFocusAPIResponse struct {
	Status string               `json:"status"`
	Data   EveningFocusResponse `json:"data"`
}
