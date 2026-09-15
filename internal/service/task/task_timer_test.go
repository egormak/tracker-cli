package task

import (
	"testing"
	"time"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/ws"
)

func TestTeaTimerModel_StatusPollMsg_UpdatesTargetDuration(t *testing.T) {
	taskTimer := &TaskTimer{
		Name:         "work",
		Role:         "work",
		TimeDuration: 25,
	}

	runningTask := entity.RunningTask{
		TaskName:       "work",
		Role:           "work",
		IsRunning:      true,
		TargetDuration: 25,
		StartTime:      time.Now().Add(-5 * time.Minute),
		Accumulated:    0,
	}

	m := newTeaTimerModel(taskTimer, runningTask)
	if m.duration != 25*time.Minute {
		t.Fatalf("expected initial duration 25m, got %v", m.duration)
	}

	// Simulate server returning status with target_duration adjusted to 30 (+5m from web)
	updatedTask := entity.RunningTask{
		TaskName:       "work",
		Role:           "work",
		IsRunning:      true,
		TargetDuration: 30,
		StartTime:      runningTask.StartTime,
		Accumulated:    0,
	}

	newModel, _ := m.Update(statusPollMsg{task: updatedTask})
	tm, ok := newModel.(teaTimerModel)
	if !ok {
		t.Fatalf("expected teaTimerModel, got %T", newModel)
	}

	if tm.duration != 30*time.Minute {
		t.Errorf("expected updated duration 30m, got %v", tm.duration)
	}
	if tm.task.TimeDuration != 30 {
		t.Errorf("expected task.TimeDuration 30, got %d", tm.task.TimeDuration)
	}
}

func TestTeaTimerModel_TogglePauseResultMsg_UpdatesTargetDuration(t *testing.T) {
	taskTimer := &TaskTimer{
		Name:         "learn",
		Role:         "learn",
		TimeDuration: 20,
	}

	runningTask := entity.RunningTask{
		TaskName:       "learn",
		Role:           "learn",
		IsRunning:      true,
		TargetDuration: 20,
		StartTime:      time.Now().Add(-2 * time.Minute),
		Accumulated:    0,
	}

	m := newTeaTimerModel(taskTimer, runningTask)

	pausedTask := entity.RunningTask{
		TaskName:       "learn",
		Role:           "learn",
		IsRunning:      false,
		TargetDuration: 25, // adjusted while pausing
		Accumulated:    2,
	}

	newModel, _ := m.Update(togglePauseResultMsg{task: pausedTask})
	tm, ok := newModel.(teaTimerModel)
	if !ok {
		t.Fatalf("expected teaTimerModel, got %T", newModel)
	}

	if tm.duration != 25*time.Minute {
		t.Errorf("expected updated duration 25m, got %v", tm.duration)
	}
	if tm.task.TimeDuration != 25 {
		t.Errorf("expected task.TimeDuration 25, got %d", tm.task.TimeDuration)
	}
	if tm.isRunning {
		t.Errorf("expected isRunning false, got true")
	}
}

func TestTeaTimerModel_GetRemainingTime(t *testing.T) {
	taskTimer := &TaskTimer{
		Name:         "work",
		Role:         "work",
		TimeDuration: 30,
	}

	runningTask := entity.RunningTask{
		TaskName:       "work",
		Role:           "work",
		IsRunning:      false,
		TargetDuration: 30,
		Accumulated:    10,
	}

	m := newTeaTimerModel(taskTimer, runningTask)
	m.elapsed = 10 * time.Minute

	remaining := m.getRemainingTime()
	if remaining != 20*time.Minute {
		t.Errorf("expected remaining 20m, got %v", remaining)
	}

	// Overtime test: elapsed > duration should clamp remaining to 0
	m.elapsed = 35 * time.Minute
	remaining = m.getRemainingTime()
	if remaining != 0 {
		t.Errorf("expected remaining clamped to 0, got %v", remaining)
	}
}

func TestTeaTimerModel_WSEventMsg_TaskAdjusted(t *testing.T) {
	taskTimer := &TaskTimer{
		Name:         "work",
		Role:         "work",
		TimeDuration: 25,
	}

	runningTask := entity.RunningTask{
		TaskName:       "work",
		Role:           "work",
		IsRunning:      true,
		TargetDuration: 25,
		StartTime:      time.Now().Add(-5 * time.Minute),
	}

	m := newTeaTimerModel(taskTimer, runningTask)

	wsMsg := wsEventMsg{
		event: ws.Event{
			Type:     ws.EventTaskAdjusted,
			TaskName: "work",
			Duration: 30,
		},
	}

	newModel, _ := m.Update(wsMsg)
	tm, ok := newModel.(teaTimerModel)
	if !ok {
		t.Fatalf("expected teaTimerModel, got %T", newModel)
	}

	if tm.duration != 30*time.Minute {
		t.Errorf("expected duration 30m, got %v", tm.duration)
	}
	if tm.task.TimeDuration != 30 {
		t.Errorf("expected task.TimeDuration 30, got %d", tm.task.TimeDuration)
	}
}

func TestTeaTimerModel_WSEventMsg_TaskPausedAndResumed(t *testing.T) {
	taskTimer := &TaskTimer{
		Name:         "english",
		Role:         "learn",
		TimeDuration: 20,
	}

	runningTask := entity.RunningTask{
		TaskName:       "english",
		Role:           "learn",
		IsRunning:      true,
		TargetDuration: 20,
		StartTime:      time.Now().Add(-5 * time.Minute),
	}

	m := newTeaTimerModel(taskTimer, runningTask)

	// Test Pause event from another client
	pauseMsg := wsEventMsg{
		event: ws.Event{
			Type:     ws.EventTaskPaused,
			TaskName: "english",
		},
	}
	newModel, _ := m.Update(pauseMsg)
	tm, _ := newModel.(teaTimerModel)
	if tm.isRunning {
		t.Errorf("expected isRunning false after pause event, got true")
	}

	// Test Resume event from another client
	resumeMsg := wsEventMsg{
		event: ws.Event{
			Type:     ws.EventTaskResumed,
			TaskName: "english",
		},
	}
	newModel2, _ := tm.Update(resumeMsg)
	tm2, _ := newModel2.(teaTimerModel)
	if !tm2.isRunning {
		t.Errorf("expected isRunning true after resume event, got false")
	}
}

