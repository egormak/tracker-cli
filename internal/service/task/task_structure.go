package task

import "time"

type TaskTimer struct {
	Name                 string
	Role                 string
	TimeDuration         int
	TimeBegin            time.Time
	TimeEnd              time.Time
	TimeDone             float64
	Percent              int
	MsgID                int
	SourceDay            string // Optional: day this task is from (for rollover tasks)
	RampStep             int    // Warm-up ramp step if duration was determined by ramp
	Origin               string // How the task was chosen ("evening", "combo"); sent to the server on start
	restLimitActive      bool
	completed            bool // set by finalizeSession when the run reached its full duration
	skipCompletionNotify bool // caller sends its own desktop notification
}
