package entity

type WeeklyDigestRole struct {
	Role         string `json:"role"`
	Planned      int    `json:"planned"`
	Done         int    `json:"done"`
	SharePercent int    `json:"share_percent"`
}

type WeeklyDigestTask struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Planned      int    `json:"planned"`
	Done         int    `json:"done"`
	PlanPercent  int    `json:"plan_percent"`
	SharePercent int    `json:"share_percent"`
}

type WeeklyDigestDelta struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Planned int    `json:"planned"`
	Done    int    `json:"done"`
	Delta   int    `json:"delta"`
}

type WeeklyDigestDay struct {
	Day       string `json:"day"`
	Date      string `json:"date"`
	TotalDone int    `json:"total_done"`
}

// WeeklyDigest mirrors GET /api/v1/stats/weekly/digest.
type WeeklyDigest struct {
	Week          string              `json:"week"`
	WeekStartDate string              `json:"week_start_date"`
	WeekEndDate   string              `json:"week_end_date"`
	TotalPlanned  int                 `json:"total_planned"`
	TotalDone     int                 `json:"total_done"`
	Roles         []WeeklyDigestRole  `json:"roles"`
	Strict        []WeeklyDigestTask  `json:"strict"`
	Free          []WeeklyDigestTask  `json:"free"`
	FreeDone      int                 `json:"free_done"`
	OverPlan      []WeeklyDigestDelta `json:"over_plan"`
	Untouched     []WeeklyDigestDelta `json:"untouched"`
	Unscheduled   []WeeklyDigestDelta `json:"unscheduled"`
	Days          []WeeklyDigestDay   `json:"days"`
	BusiestDay    *WeeklyDigestDay    `json:"busiest_day,omitempty"`
}
