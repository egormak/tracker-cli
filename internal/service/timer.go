package service

import (
	"log/slog"
	"tracker_cli/internal/repository/api"
)

func TimeDurationGet(taskName string) int {

	slog.Info("Get Time Duration")
	timeDurationDefault := api.TimeDurationGet()
	// taskScheduleTime := api.GetTaskParams(taskName).TimeDuration

	// if taskParams != (task_params.TaskParams{}) {
	// 	taskTimeDone := statistic.StatisticTaskGet(t.Name)
	// 	taskTimeLeft := (taskParams.Time*t.Procent)/100 - taskTimeDone
	// 	if taskTimeLeft <= 0 {
	// 		slog.Info("Time for this task was done")
	// 		os.Exit(0)
	// 	}

	// 	if t.TimeDuration > taskTimeLeft {
	// 		t.TimeDuration = taskTimeLeft
	// 	}
	// }

	return timeDurationDefault
}
