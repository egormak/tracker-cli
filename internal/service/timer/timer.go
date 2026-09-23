package timer

import (
	"tracker_cli/internal/repository/api"
)

// TimeListSet sets the countdown timer duration on the backend.
func TimeListSet(timerCount int) error {
	return api.TimeListSet(timerCount)
}

// SetGlobalTime sets the global scheduler timer on the backend.
func SetGlobalTime(timeScheduler int) error {
	return api.SetGlobalTime(timeScheduler)
}

// TimeDurationGet retrieves the current countdown timer duration.
func TimeDurationGet() int {
	return api.TimeDurationGet()
}

// TimeDurationDel decreases or resets the countdown timer duration on the backend.
func TimeDurationDel(timerCount int) error {
	return api.TimeDurationDel(timerCount)
}
