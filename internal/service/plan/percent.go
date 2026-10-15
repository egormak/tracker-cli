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

	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service"
	"tracker_cli/internal/service/task"
	"tracker_cli/internal/service/telegram"
)

var (
	telegramSender     = telegram.TelegramMessageSend
	planPercentRotator = api.RotatePlanPercent

	percentTaskSelector = func(useSchedule bool) (name string, percent int, duration int, sourceDay string, err error) {
		if useSchedule {
			var timeLeft int
			name, percent, timeLeft, sourceDay, err = api.GetTaskByPercentPlanSchedule()
			if err != nil {
				return "", 0, 0, "", err
			}
			defaultDuration := service.TimeDurationGet(name)
			if timeLeft > 0 && timeLeft < defaultDuration {
				duration = timeLeft
			} else {
				duration = defaultDuration
			}
			return name, percent, duration, sourceDay, nil
		}

		name, percent, err = api.GetTaskByPercentPlan()
		if err != nil {
			return "", 0, 0, "", err
		}
		duration = service.TimeDurationGet(name)
		return name, percent, duration, "", nil
	}

	percentTimerRunner = func(ctx context.Context, delay time.Duration, name string, duration int, percent int, sourceDay string) (int, error) {
		timer, err := task.CreateTaskTimerWithSourceDay(name, duration, percent, sourceDay)
		if err != nil {
			if errors.Is(err, task.ErrTaskCompleted) {
				return 0, task.ErrTaskCompleted
			}
			return 0, fmt.Errorf("initialise task timer: %w", err)
		}
		timer.SourceDay = sourceDay

		slog.Info("starting planned task", "task", name, "percent", percent, "duration", duration)

		if delay > 0 {
			slog.Info("waiting before task start", "delay_seconds", int(delay.Seconds()))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				slog.Info("interrupted during delay")
				return 0, task.ErrTaskAborted
			}
		}

		if err := timer.Run(); err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				return int(timer.TimeDone), task.ErrTaskAborted
			}
			return int(timer.TimeDone), fmt.Errorf("run task timer: %w", err)
		}

		return int(timer.TimeDone), nil
	}
)

// RunPercent runs the next task from the percent plan queue.
func RunPercent(delay time.Duration) error {
	return runPercentInternal(delay, 0, false)
}

// RunPercentSchedule runs the next task from the percent plan queue with schedule awareness.
func RunPercentSchedule(delay time.Duration) error {
	return runPercentInternal(delay, 0, true)
}

// RunPercentBatch keeps running tasks from the percent plan queue until batchDuration is used up.
func RunPercentBatch(delay time.Duration, batchDuration time.Duration, useSchedule bool) error {
	return runPercentInternal(delay, batchDuration, useSchedule)
}

func runPercentInternal(delay time.Duration, batchDuration time.Duration, useSchedule bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := interruptLogger(ctx)

	if batchDuration > 0 {
		if batchDuration < time.Minute {
			return fmt.Errorf("batch duration must be at least 1 minute")
		}

		totalBatchMin := int(batchDuration.Minutes())
		elapsedMinutes := 0
		consecutiveCompleted := 0
		const maxRotations = 10

		for elapsedMinutes < totalBatchMin {
			if logger() {
				return nil
			}

			remainingBatch := totalBatchMin - elapsedMinutes
			if remainingBatch <= 0 {
				break
			}

			actualTaskDone, err := runPercentOnce(ctx, delay, useSchedule, remainingBatch)
			if err != nil {
				if errors.Is(err, task.ErrTaskAborted) {
					return nil
				}
				if errors.Is(err, task.ErrTaskCompleted) {
					slog.Info("task plan already completed, rotating to next task")
					consecutiveCompleted++
					if consecutiveCompleted >= maxRotations {
						slog.Info("all tasks in plan completed or max rotations reached", "rotations", consecutiveCompleted)
						break
					}
					if _, rotErr := planPercentRotator(); rotErr != nil {
						slog.Warn("failed to rotate plan percent", "error", rotErr)
					}
					continue
				}
				return err
			}

			if actualTaskDone <= 0 {
				slog.Info("no task duration executed, rotating to next task", "task_done", actualTaskDone)
				consecutiveCompleted++
				if consecutiveCompleted >= maxRotations {
					break
				}
				if _, rotErr := planPercentRotator(); rotErr != nil {
					slog.Warn("failed to rotate plan percent", "error", rotErr)
				}
				continue
			}

			consecutiveCompleted = 0
			elapsedMinutes += actualTaskDone
			slog.Info("batch progress", "elapsed_min", elapsedMinutes, "total_batch_min", totalBatchMin)
		}

		if elapsedMinutes >= totalBatchMin {
			msg := fmt.Sprintf("🎉 Batch session completed! Total time: %d min.", elapsedMinutes)
			slog.Info("batch session completed", "elapsed_min", elapsedMinutes, "total_batch_min", totalBatchMin)
			telegramSender(msg)
		}
		return nil
	}

	if logger() {
		return nil
	}
	if _, err := runPercentOnce(ctx, delay, useSchedule); err != nil {
		if errors.Is(err, task.ErrTaskAborted) {
			return nil
		}
		return err
	}
	return nil
}

func runPercentOnce(ctx context.Context, delay time.Duration, useSchedule bool, maxDuration ...int) (int, error) {
	if delay < 0 {
		delay = 0
	}

	name, percent, duration, sourceDay, err := percentTaskSelector(useSchedule)
	if err != nil {
		return 0, fmt.Errorf("select task: %w", err)
	}

	if sourceDay != "" {
		slog.Info("schedule-aware task selected (rollover)", "task", name, "percent", percent, "source_day", sourceDay)
	} else {
		slog.Info("task selected", "task", name, "percent", percent)
	}

	if len(maxDuration) > 0 && maxDuration[0] > 0 && duration > maxDuration[0] {
		slog.Info("clamping task duration to remaining batch", "task", name, "requested", duration, "clamped", maxDuration[0])
		duration = maxDuration[0]
	}

	return percentTimerRunner(ctx, delay, name, duration, percent, sourceDay)
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
