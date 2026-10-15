package entity

// DebtItem is one task's shortfall on one previous day of this week.
type DebtItem struct {
	TaskName  string `json:"task_name"`
	Role      string `json:"role"`
	Day       string `json:"day"`
	Date      string `json:"date"`
	Remaining int    `json:"remaining"`
}

// DebtDay is one cell of the week strip; Status is closed, partial, open, today, future or none.
type DebtDay struct {
	Day       string `json:"day"`
	Date      string `json:"date"`
	Planned   int    `json:"planned"`
	Remaining int    `json:"remaining"`
	Status    string `json:"status"`
}

// DebtSummary is GET /api/v1/debt: the previous days' unfinished schedule of this week.
type DebtSummary struct {
	TotalMinutes   int        `json:"total_minutes"`
	Items          []DebtItem `json:"items"`
	OldestDay      string     `json:"oldest_day"`
	TodayRemaining int        `json:"today_remaining"`
	Days           []DebtDay  `json:"days"`
}

// SplitPart is one destination of a stopped timer: kind today, debt or tomorrow.
type SplitPart struct {
	Kind    string `json:"kind"`
	Day     string `json:"day"`
	Date    string `json:"date"`
	Minutes int    `json:"minutes"`
}
