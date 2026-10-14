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

// WeeklyDigestClient is the time logged through one client ("unknown" = no client name).
type WeeklyDigestClient struct {
	Client       string `json:"client"`
	Records      int    `json:"records"`
	Minutes      int    `json:"minutes"`
	SharePercent int    `json:"share_percent"`
}

type WeeklyDigestUsageClient struct {
	Client string `json:"client"`
	Count  int    `json:"count"`
}

// WeeklyDigestUsage is how many times one kind of choice was made during the week.
type WeeklyDigestUsage struct {
	Event   string                    `json:"event"`
	Count   int                       `json:"count"`
	Clients []WeeklyDigestUsageClient `json:"clients"`
}

// WeeklyDigest mirrors GET /api/v1/stats/weekly/digest. Source is "snapshot" when the
// week's records are already wiped and the server answers from the copy taken before.
type WeeklyDigest struct {
	Week            string              `json:"week"`
	Source          string              `json:"source"`
	SnapshotTakenAt string              `json:"snapshot_taken_at,omitempty"`
	WeekStartDate   string              `json:"week_start_date"`
	WeekEndDate     string              `json:"week_end_date"`
	TotalPlanned    int                 `json:"total_planned"`
	TotalDone       int                 `json:"total_done"`
	Roles           []WeeklyDigestRole  `json:"roles"`
	Strict          []WeeklyDigestTask  `json:"strict"`
	Free            []WeeklyDigestTask  `json:"free"`
	FreeDone        int                 `json:"free_done"`
	OverPlan        []WeeklyDigestDelta `json:"over_plan"`
	Untouched       []WeeklyDigestDelta `json:"untouched"`
	Unscheduled     []WeeklyDigestDelta `json:"unscheduled"`
	Days            []WeeklyDigestDay   `json:"days"`
	BusiestDay      *WeeklyDigestDay    `json:"busiest_day,omitempty"`

	Clients []WeeklyDigestClient `json:"clients"`
	Usage   []WeeklyDigestUsage  `json:"usage"`
}
