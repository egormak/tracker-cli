package task

import (
	"time"

	"tracker_cli/internal/domain/entity"
)

type TaskTimer struct {
	Name                 string
	Role                 string
	TimeDuration         int
	TimeBegin            time.Time
	TimeEnd              time.Time
	TimeDone             float64
	Percent              int
	MsgID                int
	SourceDay            string             // Optional: day this task is from (for rollover tasks)
	RampStep             int                // Warm-up ramp step if duration was determined by ramp
	Origin               string             // How the task was chosen ("evening", "combo", "batch"); sent to the server on start
	DebtShare            int                // 0–100: % of the session that pays the task's past debt on stop; sent on start
	Split                []entity.SplitPart // where the stopped session's minutes went (from the server)
	completed            bool               // set by finalizeSession when the run reached its full duration
	skipCompletionNotify bool               // caller sends its own desktop notification
}
