package statistic

import (
	"log/slog"
	"math"
	"time"
	"tracker_cli/internal/repository/api"
)

// ShowAfterTask logs the task's result and the day's progress from the dashboard state.
func ShowAfterTask(taskName string) {
	state, err := api.GetDashboardState()
	if err != nil {
		slog.Error("failed to fetch dashboard state", "error", err)
		return
	}

	t, ok := state.Task(taskName)
	if ok && t.TimeDuration > 0 {
		slog.Info("Show Result", "task", taskName, "time_duration", t.TimeDone, "left", t.TimeDuration-t.TimeDone)
	} else {
		slog.Info("Show Result", "task", taskName, "time_duration", t.TimeDone)
	}

	if state.TotalPlanned == 0 {
		slog.Info("Time Limit for Today is not Set")
		return
	}

	completionPercentage := statCompletionPercentage(state.TotalDone, state.TotalPlanned)
	timeLeft := state.TotalPlanned - state.TotalDone
	timePrediction := time.Now().Add(time.Minute * time.Duration(timeLeft))

	slog.Info("Percent Done", "percent", completionPercentage)
	slog.Info("Time Done", "time", state.TotalDone)
	slog.Info("Time prediction", "time", timePrediction.Format("15:04:05"))
}

func statCompletionPercentage(completionTimeDone, scheduledTimeToday int) float64 {
	percent := float64(completionTimeDone) / float64(scheduledTimeToday) * 100
	round := math.Round(percent*100) / 100
	return round
}
