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

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/pkg/restutil"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/task"
	"tracker_cli/internal/service/telegram"
	"tracker_cli/internal/service/timer"
)

// RunBacklog triggers the gamified / backlog sequence.
func RunBacklog(delay time.Duration, restLimitMinutes int) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := interruptLogger(ctx)

	for {
		if logger() {
			return nil
		}

		// Fetch all rollover tasks on each iteration to get updated RemainingTimes
		rollovers, err := api.GetRolloverTasks()
		if err != nil {
			return fmt.Errorf("fetch rollover tasks: %w", err)
		}

		if len(rollovers) == 0 {
			slog.Info("no backlog tasks found or all completed")
			return nil
		}

		workDone := false

		for _, rollover := range rollovers {
			if logger() {
				return nil
			}

			if rollover.RemainingTime <= 0 {
				continue
			}

			if restLimitMinutes >= 0 {
				limitUnits := restutil.UnitsFromMinutes(restLimitMinutes)
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
			}

			restLimitActive := restLimitMinutes >= 0

			if err := runBacklogOnce(ctx, delay, restLimitActive, rollover); err != nil {
				if errors.Is(err, task.ErrTaskAborted) {
					return nil
				}
				return err
			}

			workDone = true
		}

		// If we looped through all rollovers and none had RemainingTime > 0, we are done
		if !workDone {
			slog.Info("all backlog tasks completed")
			break
		}
	}

	return nil
}

func runBacklogOnce(ctx context.Context, delay time.Duration, restLimitActive bool, rollover entity.RolloverTask) error {
	if delay < 0 {
		delay = 0
	}

	if rollover.RemainingTime <= 0 {
		slog.Info("skipping task, remaining time is 0", "task", rollover.TaskName)
		return nil
	}

	// Fetch next duration from time list
	timeDuration := timer.TimeDurationGet()

	// Ensure we don't run longer than the deficit
	if timeDuration > rollover.RemainingTime {
		slog.Info("capping duration to remaining time", "next_duration", timeDuration, "remaining", rollover.RemainingTime)
		timeDuration = rollover.RemainingTime
	}

	slog.Info("backlog task selected", "task", rollover.TaskName, "source_day", rollover.SourceDay, "duration", timeDuration, "remaining", rollover.RemainingTime)

	// In the CLI, percent is usually passed. RolloverTask has Percent. We use it.
	timerObj, err := task.CreateTaskTimer(rollover.TaskName, timeDuration, rollover.Percent)
	if err != nil {
		if errors.Is(err, task.ErrTaskCompleted) {
			slog.Info("task plan already completed", "task", rollover.TaskName)
			return nil
		}
		return fmt.Errorf("initialise task timer: %w", err)
	}
	
	timerObj.SourceDay = rollover.SourceDay
	timerObj.SetRestLimitActive(restLimitActive)

	slog.Info("starting backlog task", "task", rollover.TaskName, "duration", timeDuration)

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

	if err := timerObj.Run(); err != nil {
		if errors.Is(err, task.ErrTaskAborted) {
			return task.ErrTaskAborted
		}
		return fmt.Errorf("run task timer: %w", err)
	}

	return nil
}
