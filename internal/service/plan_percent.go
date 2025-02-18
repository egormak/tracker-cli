package service

import (
	"fmt"
	"log/slog"
	"time"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/task"
)

func PlanPercent() {

	// Get Params
	name, percent := api.GetTaskByPercentPlan()
	timeDuration := TimeDurationGet(name)
	taskApp := task.CreateTaskTimer(name, timeDuration, percent)

	slog.Info(fmt.Sprintf("\033[33mTaskName:\033[32m %s\033[0m", name), "percent", percent)
	time.Sleep(time.Second * 15)

	taskApp.Run()

}
