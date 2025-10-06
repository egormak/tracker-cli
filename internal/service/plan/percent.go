package plan

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service"
	"tracker_cli/internal/service/task"
)

// RunPercent triggers the next task from the percent plan queue.
func RunPercent(delay time.Duration) error {
	if delay < 0 {
		delay = 0
	}

	name, percent, err := api.GetTaskByPercentPlan()
	if err != nil {
		return fmt.Errorf("fetch next planned task: %w", err)
	}

	timeDuration := service.TimeDurationGet(name)
	timer, err := task.CreateTaskTimer(name, timeDuration, percent)
	if err != nil {
		if errors.Is(err, task.ErrTaskCompleted) {
			slog.Info("task plan already completed", "task", name)
			return nil
		}
		return fmt.Errorf("initialise task timer: %w", err)
	}

	slog.Info("starting planned task", "task", name, "percent", percent, "duration", timeDuration)

	if delay > 0 {
		time.Sleep(delay)
	}

	if err := timer.Run(); err != nil {
		return fmt.Errorf("run task timer: %w", err)
	}

	return nil
}
