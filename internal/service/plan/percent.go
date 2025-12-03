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
	"tracker_cli/internal/service/telegram"
)

// RunPercent triggers the next task from the percent plan queue.
func RunPercent(delay time.Duration, restLimitMinutes int) error {
	return runPercentInternal(delay, restLimitMinutes, false)
}

// RunPercentSchedule triggers the next task from the percent plan queue with schedule awareness.
func RunPercentSchedule(delay time.Duration, restLimitMinutes int) error {
	return runPercentInternal(delay, restLimitMinutes, true)
}

func runPercentInternal(delay time.Duration, restLimitMinutes int, useSchedule bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := interruptLogger(ctx)

	if restLimitMinutes < 0 {
		if logger() {
			return nil
		}
		if err := runPercentOnce(ctx, delay, false, useSchedule); err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				return nil
			}
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
			telegram.TelegramMessageSend(fmt.Sprintf("Rest limit reached: %.1f minutes available (limit %d). Take a break or do some exercise.", currentMinutes, restLimitMinutes))
			return nil
		}

		if err := runPercentOnce(ctx, delay, true, useSchedule); err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				return nil
			}
			return err
		}
	}
}

func runPercentOnce(ctx context.Context, delay time.Duration, restLimitActive bool, useSchedule bool) error {
	if delay < 0 {
		delay = 0
	}

	var name string
	var percent int
	var sourceDay string
	var timeDuration int
	var err error

	if useSchedule {
		var timeLeft int
		name, percent, timeLeft, sourceDay, err = api.GetTaskByPercentPlanSchedule()
		if err != nil {
			return fmt.Errorf("fetch next scheduled task: %w", err)
		}
		if sourceDay != "" {
			slog.Info("schedule-aware task selected (rollover)", "task", name, "percent", percent, "time_left", timeLeft, "source_day", sourceDay)
		} else {
			slog.Info("schedule-aware task selected", "task", name, "percent", percent, "time_left", timeLeft)
		}
		// Use the smaller value between timeLeft and default timer duration
		defaultDuration := service.TimeDurationGet(name)
		if timeLeft > 0 && timeLeft < defaultDuration {
			timeDuration = timeLeft
		} else {
			timeDuration = defaultDuration
		}
	} else {
		name, percent, err = api.GetTaskByPercentPlan()
		if err != nil {
			return fmt.Errorf("fetch next planned task: %w", err)
		}
		timeDuration = service.TimeDurationGet(name)
	}

	timer, err := task.CreateTaskTimer(name, timeDuration, percent)
	if err != nil {
		if errors.Is(err, task.ErrTaskCompleted) {
			slog.Info("task plan already completed", "task", name)
			return nil
		}
		return fmt.Errorf("initialise task timer: %w", err)
	}
	timer.SourceDay = sourceDay
	timer.SetRestLimitActive(restLimitActive)

	slog.Info("starting planned task", "task", name, "percent", percent, "duration", timeDuration)

	// Use context-aware sleep so CTRL+C can interrupt during the delay
	if delay > 0 {
		slog.Info("waiting before task start", "delay_seconds", int(delay.Seconds()))
		select {
		case <-time.After(delay):
			// Delay completed normally
		case <-ctx.Done():
			// Interrupted during delay
			slog.Info("interrupted during delay")
			return task.ErrTaskAborted
		}
	}

	if err := timer.Run(); err != nil {
		if errors.Is(err, task.ErrTaskAborted) {
			return task.ErrTaskAborted
		}
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
