package plan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tracker_cli/internal/pkg/restutil"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service"
	"tracker_cli/internal/service/task"
)

// RunPercent triggers the next task from the percent plan queue.
func RunPercent(delay time.Duration, restLimitMinutes int) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := interruptLogger(ctx)

	if restLimitMinutes < 0 {
		if logger() {
			return nil
		}
		if err := runPercentOnce(delay); err != nil {
			return err
		}
		if logger() {
			return nil
		}
		return nil
	}

	limitUnits := restutil.UnitsFromMinutes(restLimitMinutes)

	for {
		if logger() {
			return nil
		}

		restUnits, err := api.GetRestTime()
		if err != nil {
			return fmt.Errorf("fetch rest balance: %w", err)
		}

		currentMinutes := restutil.MinutesFromUnits(restUnits)
		if restUnits > limitUnits {
			slog.Info("rest limit reached", "rest_minutes", currentMinutes, "limit_minutes", restLimitMinutes)
			return nil
		}

		if err := runPercentOnce(delay); err != nil {
			return err
		}
	}
}

func runPercentOnce(delay time.Duration) error {
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

func interruptLogger(ctx context.Context) func() bool {
	var logged bool

	return func() bool {
		if ctx.Err() == nil {
			return false
		}

		if !logged {
			slog.Info("plan percent interrupted")
			logged = true
		}

		return true
	}
}
