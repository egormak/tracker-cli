package entity

type Timers struct {
	TimeDuration int `json:"time_duration"`
	Count        int `json:"count"`
}

func (t Timers) Duration() int {
	if t.TimeDuration > 0 {
		return t.TimeDuration
	}
	return t.Count
}
