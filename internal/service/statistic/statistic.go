package statistic

import (
	"log/slog"
	"math"
	"time"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/task_params"
)

func StatisticTaskShow(taskName string) {
	taskResult := StatisticTaskGet(taskName)
	timeSchedule := task_params.GetTaskParams(taskName).Time

	if timeSchedule == 0 {
		slog.Info("Show Result", "task", taskName, "time_duration", taskResult)
	} else {
		slog.Info("Show Result", "task", taskName, "time_duration", taskResult, "left", timeSchedule-taskResult)
	}
}

func StatisticTaskGet(taskName string) int {
	return api.StatisticTaskGet(taskName)
}

func StatisticFullShow() {
	scheduledTimeToday, err := api.GetScheduledTimeToday()
	if err != nil || scheduledTimeToday == 0 {
		slog.Info("Time Limit for Today is not Set")
		return
	}

	completionTimeDone, err := api.GetCompletionTimeDoneToday()
	if err != nil {
		slog.Error("Failed to fetch completion time for today", "error", err)
		return
	}

	completionPercentage := statCompletionPercentage(completionTimeDone, scheduledTimeToday)
	timeLeft := scheduledTimeToday - completionTimeDone
	timePrediction := time.Now().Add(time.Minute * time.Duration(timeLeft))

	slog.Info("Percent Done", "percent", completionPercentage)
	slog.Info("Time Done", "time", completionTimeDone)
	slog.Info("Time prediction", "time", timePrediction.Format("15:04:05"))
}

func statCompletionPercentage(completionTimeDone, scheduledTimeToday int) float64 {
	percent := float64(completionTimeDone) / float64(scheduledTimeToday) * 100
	round := math.Round(percent*100) / 100
	return round
}

