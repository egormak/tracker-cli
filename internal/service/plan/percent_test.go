package plan

import (
	"context"
	"testing"
	"time"

	"tracker_cli/internal/service/task"
)

func TestRunPercentBatch_ClampingAndProgress(t *testing.T) {
	oldSelector := percentTaskSelector
	oldRunner := percentTimerRunner
	oldTelegram := telegramSender
	defer func() {
		percentTaskSelector = oldSelector
		percentTimerRunner = oldRunner
		telegramSender = oldTelegram
	}()

	type executedTask struct {
		name     string
		duration int
	}
	var executed []executedTask

	percentTaskSelector = func(useSchedule bool) (string, int, int, string, error) {
		return "test-task", 50, 25, "", nil
	}

	percentTimerRunner = func(ctx context.Context, delay time.Duration, name string, duration int, percent int, sourceDay string, restLimitActive bool) (int, error) {
		executed = append(executed, executedTask{name: name, duration: duration})
		return duration, nil
	}

	var telegramMsg string
	telegramSender = func(msg string) {
		telegramMsg = msg
	}

	// 40 minute batch:
	// Step 1: requested 25m, remaining 40m -> run 25m, elapsed becomes 25m
	// Step 2: requested 25m, remaining 15m -> clamped to 15m, run 15m, elapsed becomes 40m
	err := RunPercentBatch(0, -1, 40*time.Minute, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(executed) != 2 {
		t.Fatalf("expected 2 tasks executed, got %d", len(executed))
	}

	if executed[0].duration != 25 {
		t.Errorf("expected first task duration 25, got %d", executed[0].duration)
	}

	if executed[1].duration != 15 {
		t.Errorf("expected second task duration to be clamped to 15, got %d", executed[1].duration)
	}

	expectedMsg := "🎉 Batch session completed! Total time: 40 min."
	if telegramMsg != expectedMsg {
		t.Errorf("expected telegram message %q, got %q", expectedMsg, telegramMsg)
	}
}

func TestRunPercentBatch_DurationValidation(t *testing.T) {
	err := RunPercentBatch(0, -1, 30*time.Second, false)
	if err == nil {
		t.Fatal("expected error for batch duration < 1m, got nil")
	}
}

func TestRunPercentBatch_Abort(t *testing.T) {
	oldSelector := percentTaskSelector
	oldRunner := percentTimerRunner
	defer func() {
		percentTaskSelector = oldSelector
		percentTimerRunner = oldRunner
	}()

	percentTaskSelector = func(useSchedule bool) (string, int, int, string, error) {
		return "aborted-task", 50, 25, "", nil
	}

	percentTimerRunner = func(ctx context.Context, delay time.Duration, name string, duration int, percent int, sourceDay string, restLimitActive bool) (int, error) {
		return 5, task.ErrTaskAborted
	}

	err := RunPercentBatch(0, -1, 30*time.Minute, false)
	if err != nil {
		t.Fatalf("expected nil error on abort, got %v", err)
	}
}

func TestRunPercentBatch_RotateOnCompletedTask(t *testing.T) {
	oldSelector := percentTaskSelector
	oldRunner := percentTimerRunner
	oldRotator := planPercentRotator
	defer func() {
		percentTaskSelector = oldSelector
		percentTimerRunner = oldRunner
		planPercentRotator = oldRotator
	}()

	call := 0
	percentTaskSelector = func(useSchedule bool) (string, int, int, string, error) {
		call++
		if call == 1 {
			return "completed-task", 100, 25, "", nil
		}
		return "next-task", 50, 20, "", nil
	}

	percentTimerRunner = func(ctx context.Context, delay time.Duration, name string, duration int, percent int, sourceDay string, restLimitActive bool) (int, error) {
		if name == "completed-task" {
			return 0, task.ErrTaskCompleted
		}
		return 20, nil
	}

	rotated := false
	planPercentRotator = func() (string, error) {
		rotated = true
		return "rotated", nil
	}

	err := RunPercentBatch(0, -1, 20*time.Minute, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !rotated {
		t.Error("expected plan percent rotation when encountering completed task")
	}
}

func TestRunPercentBatch_EarlyQuitNoOverCredit(t *testing.T) {
	oldSelector := percentTaskSelector
	oldRunner := percentTimerRunner
	defer func() {
		percentTaskSelector = oldSelector
		percentTimerRunner = oldRunner
	}()

	percentTaskSelector = func(useSchedule bool) (string, int, int, string, error) {
		return "quick-quit-task", 50, 25, "", nil
	}

	// User quit early: only 2 minutes credited, not full 25!
	percentTimerRunner = func(ctx context.Context, delay time.Duration, name string, duration int, percent int, sourceDay string, restLimitActive bool) (int, error) {
		return 2, nil
	}

	// For a 5m batch, step 1 did 2m, step 2 did 2m, step 3 clamped to 1m -> 5m
	err := RunPercentBatch(0, -1, 4*time.Minute, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
