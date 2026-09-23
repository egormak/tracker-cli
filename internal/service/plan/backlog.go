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
)

var (
	backlogRolloversGetter = api.GetRolloverTasks
	backlogDurationGetter  = api.TimeDurationGet

	backlogTimerRunner = func(ctx context.Context, delay time.Duration, rollover entity.RolloverTask, duration int, restLimitActive bool) (int, error) {
		timerObj, err := task.CreateTaskTimerWithSourceDay(rollover.TaskName, duration, rollover.Percent, rollover.SourceDay)
		if err != nil {
			if errors.Is(err, task.ErrTaskCompleted) {
				slog.Info("task plan already completed", "task", rollover.TaskName)
				return 0, nil
			}
			return 0, fmt.Errorf("initialise task timer: %w", err)
		}

		timerObj.SourceDay = rollover.SourceDay
		timerObj.SetRestLimitActive(restLimitActive)

		slog.Info("starting backlog task", "task", rollover.TaskName, "duration", duration)

		if delay > 0 {
			slog.Info("waiting before task start", "delay_seconds", int(delay.Seconds()))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				slog.Info("interrupted during delay")
				return 0, task.ErrTaskAborted
			}
		}

		if err := timerObj.Run(); err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				return int(timerObj.TimeDone), task.ErrTaskAborted
			}
			return int(timerObj.TimeDone), fmt.Errorf("run task timer: %w", err)
		}

		return int(timerObj.TimeDone), nil
	}
)

// RunBacklog triggers the gamified / backlog sequence.
func RunBacklog(delay time.Duration, restLimitMinutes int, batchDuration ...time.Duration) error {
	var batch time.Duration
	if len(batchDuration) > 0 {
		batch = batchDuration[0]
	}
	return runBacklogInternal(delay, restLimitMinutes, batch)
}

// RunBacklogBatch triggers the gamified / backlog sequence for a target batch duration.
func RunBacklogBatch(delay time.Duration, restLimitMinutes int, batchDuration time.Duration) error {
	return runBacklogInternal(delay, restLimitMinutes, batchDuration)
}

func runBacklogInternal(delay time.Duration, restLimitMinutes int, batchDuration time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := interruptLogger(ctx)

	if batchDuration > 0 {
		if batchDuration < time.Minute {
			return fmt.Errorf("batch duration must be at least 1 minute")
		}

		totalBatchMin := int(batchDuration.Minutes())
		elapsedMinutes := 0

		for elapsedMinutes < totalBatchMin {
			if logger() {
				return nil
			}

			remainingBatch := totalBatchMin - elapsedMinutes
			if remainingBatch <= 0 {
				break
			}

			rollovers, err := backlogRolloversGetter()
			if err != nil {
				return fmt.Errorf("fetch rollover tasks: %w", err)
			}

			if len(rollovers) == 0 {
				slog.Info("no backlog tasks found or all completed")
				break
			}

			workDone := false

			for _, rollover := range rollovers {
				if logger() {
					return nil
				}

				remainingBatch = totalBatchMin - elapsedMinutes
				if remainingBatch <= 0 {
					break
				}

				if rollover.RemainingTime <= 0 {
					continue
				}

				if restLimitMinutes >= 0 {
					limitUnits := restutil.UnitsFromMinutes(restLimitMinutes)
					restUnits, err := restBalanceGetter()
					if err != nil {
						return fmt.Errorf("fetch rest balance: %w", err)
					}

					currentMinutes := restutil.MinutesFromUnits(restUnits)
					if restUnits > limitUnits {
						slog.Info("rest limit reached", "rest_minutes", currentMinutes, "limit_minutes", restLimitMinutes)
						telegramSender(fmt.Sprintf("Rest limit reached: %.1f minutes available (limit %d). Take a break or do some exercise.", currentMinutes, restLimitMinutes))
						return nil
					}
				}

				restLimitActive := restLimitMinutes >= 0

				actualTaskDone, err := runBacklogOnce(ctx, delay, restLimitActive, rollover, remainingBatch)
				if err != nil {
					if errors.Is(err, task.ErrTaskAborted) {
						return nil
					}
					return err
				}

				if actualTaskDone <= 0 {
					continue
				}

				elapsedMinutes += actualTaskDone
				slog.Info("batch progress", "elapsed_min", elapsedMinutes, "total_batch_min", totalBatchMin)
				workDone = true

				if elapsedMinutes >= totalBatchMin {
					break
				}
			}

			if !workDone {
				slog.Info("all backlog tasks completed")
				break
			}
		}

		if elapsedMinutes >= totalBatchMin {
			msg := fmt.Sprintf("🎉 Batch session completed! Total time: %d min.", elapsedMinutes)
			slog.Info("batch session completed", "elapsed_min", elapsedMinutes, "total_batch_min", totalBatchMin)
			telegramSender(msg)
		}
		return nil
	}

	for {
		if logger() {
			return nil
		}

		// Fetch all rollover tasks on each iteration to get updated RemainingTimes
		rollovers, err := backlogRolloversGetter()
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
				restUnits, err := restBalanceGetter()
				if err != nil {
					return fmt.Errorf("fetch rest balance: %w", err)
				}

				currentMinutes := restutil.MinutesFromUnits(restUnits)
				if restUnits > limitUnits {
					slog.Info("rest limit reached", "rest_minutes", currentMinutes, "limit_minutes", restLimitMinutes)
					telegramSender(fmt.Sprintf("Rest limit reached: %.1f minutes available (limit %d). Take a break or do some exercise.", currentMinutes, restLimitMinutes))
					return nil
				}
			}

			restLimitActive := restLimitMinutes >= 0

			if _, err := runBacklogOnce(ctx, delay, restLimitActive, rollover); err != nil {
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

func runBacklogOnce(ctx context.Context, delay time.Duration, restLimitActive bool, rollover entity.RolloverTask, maxDuration ...int) (int, error) {
	if delay < 0 {
		delay = 0
	}

	if rollover.RemainingTime <= 0 {
		slog.Info("skipping task, remaining time is 0", "task", rollover.TaskName)
		return 0, nil
	}

	// Fetch next duration from time list
	timeDuration := backlogDurationGetter()

	// Ensure we don't run longer than the deficit
	if timeDuration > rollover.RemainingTime {
		slog.Info("capping duration to remaining time", "next_duration", timeDuration, "remaining", rollover.RemainingTime)
		timeDuration = rollover.RemainingTime
	}

	// Clamp to remaining batch time if provided
	if len(maxDuration) > 0 && maxDuration[0] > 0 && timeDuration > maxDuration[0] {
		slog.Info("capping duration to remaining batch time", "duration", timeDuration, "remaining_batch", maxDuration[0])
		timeDuration = maxDuration[0]
	}

	slog.Info("backlog task selected", "task", rollover.TaskName, "source_day", rollover.SourceDay, "duration", timeDuration, "remaining", rollover.RemainingTime)

	return backlogTimerRunner(ctx, delay, rollover, timeDuration, restLimitActive)
}
