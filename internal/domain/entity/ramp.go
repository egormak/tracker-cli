package entity

type RampConfig struct {
	CapMinutes             int      `json:"cap_minutes"`
	EnabledRoles           []string `json:"enabled_roles"`
	EnabledTasks           []string `json:"enabled_tasks"`
	ExcludedTasks          []string `json:"excluded_tasks"`
	DefaultFallbackMinutes int      `json:"default_fallback_minutes"`
}

type RampStatus struct {
	CurrentStep       int  `json:"current_step"`
	CapMinutes        int  `json:"cap_minutes"`
	IsCapped          bool `json:"is_capped"`
	TodayFocusMinutes int  `json:"today_focus_minutes"`
	// ResetBaselineMinutes is today's focus at the last manual reset (0 if none today);
	// EffectiveFocusMinutes is the focus counted since then, which drives CurrentStep.
	ResetBaselineMinutes  int        `json:"reset_baseline_minutes"`
	EffectiveFocusMinutes int        `json:"effective_focus_minutes"`
	Date                  string     `json:"date"`
	Config                RampConfig `json:"config"`
}
