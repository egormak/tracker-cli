package plan

import (
	"context"
	"testing"
	"time"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/service/task"
)

func TestRunBacklogBatch_ClampingAndProgress(t *testing.T) {
	oldRollovers := backlogRolloversGetter
	oldDuration := backlogDurationGetter
	oldRunner := backlogTimerRunner
	oldTelegram := telegramSender
	defer func() {
		backlogRolloversGetter = oldRollovers
		backlogDurationGetter = oldDuration
		backlogTimerRunner = oldRunner
		telegramSender = oldTelegram
	}()

	type executedTask struct {
		name     string
		duration int
	}
	var executed []executedTask

	backlogRolloversGetter = func() ([]entity.RolloverTask, error) {
		return []entity.RolloverTask{
			{TaskName: "backlog-1", RemainingTime: 20, Percent: 50, SourceDay: "monday"},
			{TaskName: "backlog-2", RemainingTime: 30, Percent: 50, SourceDay: "tuesday"},
		}, nil
	}

	backlogDurationGetter = func() int {
		return 25 // Default duration
	}

	backlogTimerRunner = func(ctx context.Context, delay time.Duration, rollover entity.RolloverTask, duration int, restLimitActive bool) (int, error) {
		executed = append(executed, executedTask{name: rollover.TaskName, duration: duration})
		return duration, nil
	}

	var telegramMsg string
	telegramSender = func(msg string) {
		telegramMsg = msg
	}

	// 30 minute batch:
	// Task 1: RemainingTime is 20, default duration is 25 -> capped to RemainingTime 20. Remaining batch is 30 >= 20. Runs 20m.
	// Task 2: RemainingTime is 30, default duration is 25 -> remaining batch is 10m -> clamped to 10m. Runs 10m.
	// Total elapsed: 30m.
	err := RunBacklogBatch(0, -1, 30*time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(executed) != 2 {
		t.Fatalf("expected 2 tasks executed, got %d", len(executed))
	}

	if executed[0].duration != 20 {
		t.Errorf("expected first task duration 20, got %d", executed[0].duration)
	}

	if executed[1].duration != 10 {
		t.Errorf("expected second task duration clamped to remaining batch 10, got %d", executed[1].duration)
	}

	expectedMsg := "🎉 Batch session completed! Total time: 30 min."
	if telegramMsg != expectedMsg {
		t.Errorf("expected telegram message %q, got %q", expectedMsg, telegramMsg)
	}
}

func TestRunBacklogBatch_DurationValidation(t *testing.T) {
	err := RunBacklogBatch(0, -1, 30*time.Second)
	if err == nil {
		t.Fatal("expected error for batch duration < 1m, got nil")
	}
}

func TestRunBacklogBatch_EmptyRollovers(t *testing.T) {
	oldRollovers := backlogRolloversGetter
	defer func() { backlogRolloversGetter = oldRollovers }()

	backlogRolloversGetter = func() ([]entity.RolloverTask, error) {
		return []entity.RolloverTask{}, nil
	}

	err := RunBacklogBatch(0, -1, 30*time.Minute)
	if err != nil {
		t.Fatalf("expected no error for empty rollovers, got %v", err)
	}
}

func TestRunBacklogBatch_Abort(t *testing.T) {
	oldRollovers := backlogRolloversGetter
	oldDuration := backlogDurationGetter
	oldRunner := backlogTimerRunner
	defer func() {
		backlogRolloversGetter = oldRollovers
		backlogDurationGetter = oldDuration
		backlogTimerRunner = oldRunner
	}()

	backlogRolloversGetter = func() ([]entity.RolloverTask, error) {
		return []entity.RolloverTask{
			{TaskName: "abort-task", RemainingTime: 20, Percent: 50},
		}, nil
	}

	backlogDurationGetter = func() int { return 15 }

	backlogTimerRunner = func(ctx context.Context, delay time.Duration, rollover entity.RolloverTask, duration int, restLimitActive bool) (int, error) {
		return 5, task.ErrTaskAborted
	}

	err := RunBacklogBatch(0, -1, 30*time.Minute)
	if err != nil {
		t.Fatalf("expected nil error on abort, got %v", err)
	}
}
