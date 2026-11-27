package task

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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

// A heartbeat tick that lands while paused must keep the loop alive: before the fix it
// ended the loop, so after resume no heartbeat was sent and the server's watchdog paused
// the task again 3 minutes later.
func TestTeaTimerModel_HeartbeatSurvivesPauseResume(t *testing.T) {
	var sent []string
	orig := heartbeatSender
	heartbeatSender = func(name string) (entity.RunningTask, error) {
		sent = append(sent, name)
		return entity.RunningTask{}, nil
	}
	t.Cleanup(func() { heartbeatSender = orig })

	m := newTeaTimerModel(&TaskTimer{Name: "games", TimeDuration: 30}, entity.RunningTask{
		TaskName: "games", IsRunning: true, TargetDuration: 30, StartTime: time.Now(),
	})

	// Paused: the tick sends nothing but schedules the next one.
	m.isRunning = false
	next, cmd := m.Update(heartbeatTickMsg{})
	m = next.(teaTimerModel)
	if cmd == nil {
		t.Fatal("paused tick ended the heartbeat loop")
	}
	if len(sent) != 0 {
		t.Fatalf("paused tick sent a heartbeat: %v", sent)
	}

	// Resumed: the next tick sends a heartbeat and schedules the following one.
	next, _ = m.Update(togglePauseResultMsg{task: entity.RunningTask{TaskName: "games", IsRunning: true, TargetDuration: 30, StartTime: time.Now()}})
	m = next.(teaTimerModel)
	_, cmd = m.Update(heartbeatTickMsg{})
	batch, ok := cmd().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatalf("expected a batch of send + next tick, got %T", cmd())
	}
	batch[0]()
	if len(sent) != 1 || sent[0] != "games" {
		t.Fatalf("expected one heartbeat for games after resume, got %v", sent)
	}

	// Stopping ends the loop.
	m.stopping = true
	if _, cmd := m.Update(heartbeatTickMsg{}); cmd != nil {
		t.Fatal("expected no heartbeat while stopping")
	}
}

var errServerDown = errors.New("connection refused")

// retrySeams replaces the retry seams with a test clock, 1 ms intervals and stubs that never
// touch the network; t.Cleanup restores them.
type retrySeams struct {
	now        time.Time
	pollErr    error
	stopErr    error
	stopRecord entity.TaskRecord
}

func swapRetrySeams(t *testing.T) *retrySeams {
	t.Helper()
	s := &retrySeams{now: time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC), pollErr: errServerDown, stopErr: errServerDown}

	origWindow, origPoll, origStop := serverRetryWindow, pollInterval, stopRetryInterval
	origNow, origFetch, origStopFn := timerNow, statusFetcher, stopRunningTask
	t.Cleanup(func() {
		serverRetryWindow, pollInterval, stopRetryInterval = origWindow, origPoll, origStop
		timerNow, statusFetcher, stopRunningTask = origNow, origFetch, origStopFn
	})

	serverRetryWindow = 30 * time.Second
	pollInterval = time.Millisecond
	stopRetryInterval = time.Millisecond
	timerNow = func() time.Time { return s.now }
	statusFetcher = func(string) (entity.RunningTask, error) { return entity.RunningTask{}, s.pollErr }
	stopRunningTask = func(string) (entity.TaskRecord, error) { return s.stopRecord, s.stopErr }
	return s
}

func newRetryTestModel() teaTimerModel {
	return newTeaTimerModel(&TaskTimer{Name: "focus", Role: "work", TimeDuration: 30}, entity.RunningTask{
		TaskName: "focus", Role: "work", IsRunning: true, TargetDuration: 30, StartTime: time.Now(),
	})
}

func updateTimer(t *testing.T, m teaTimerModel, msg tea.Msg) (teaTimerModel, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	tm, ok := next.(teaTimerModel)
	if !ok {
		t.Fatalf("expected teaTimerModel, got %T", next)
	}
	return tm, cmd
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestTeaTimerModel_PollErrorWithinWindowKeepsRunning(t *testing.T) {
	s := swapRetrySeams(t)
	t0 := s.now
	m := newRetryTestModel()

	for _, at := range []time.Duration{0, 29 * time.Second} {
		s.now = t0.Add(at)
		var cmd tea.Cmd
		m, cmd = updateTimer(t, m, statusPollMsg{err: errServerDown})
		if cmd == nil || isQuit(cmd) {
			t.Fatalf("poll error at t0+%s: expected a retry, got quit or nil", at)
		}
		if m.exitState != (exitState{}) {
			t.Fatalf("poll error at t0+%s: exitState = %+v, want zero", at, m.exitState)
		}
		if got := m.getStatus(); got != "server unreachable, retrying" {
			t.Fatalf("poll error at t0+%s: status = %q", at, got)
		}
	}

	// A successful poll closes the window.
	s.now = t0.Add(29500 * time.Millisecond)
	m, _ = updateTimer(t, m, statusPollMsg{task: entity.RunningTask{
		TaskName: "focus", Role: "work", IsRunning: true, TargetDuration: 30, StartTime: time.Now(),
	}})
	if !m.pollFailingSince.IsZero() {
		t.Fatalf("successful poll left pollFailingSince = %v", m.pollFailingSince)
	}
	if got := m.getStatus(); got != "running" {
		t.Fatalf("status after a successful poll = %q, want running", got)
	}

	// A new error starts a new window.
	s.now = t0.Add(50 * time.Second)
	m, cmd := updateTimer(t, m, statusPollMsg{err: errServerDown})
	if cmd == nil || isQuit(cmd) {
		t.Fatal("poll error after a success: expected a retry, got quit or nil")
	}
	if m.exitState != (exitState{}) {
		t.Fatalf("exitState = %+v, want zero", m.exitState)
	}
}

func TestTeaTimerModel_PollErrorPastWindowExitsServerLost(t *testing.T) {
	s := swapRetrySeams(t)
	t0 := s.now
	m := newRetryTestModel()

	m, _ = updateTimer(t, m, statusPollMsg{err: errServerDown})
	s.now = t0.Add(30 * time.Second)
	m, cmd := updateTimer(t, m, statusPollMsg{err: errServerDown})
	if !isQuit(cmd) {
		t.Fatal("expected tea.Quit once the window is over")
	}
	if !m.exitState.serverLost {
		t.Fatal("expected serverLost")
	}
	if m.exitState.shouldSave {
		t.Fatal("expected shouldSave false")
	}
}

func TestTeaTimerModel_StopRetrySavesOnSuccess(t *testing.T) {
	s := swapRetrySeams(t)
	m := newRetryTestModel()
	m.task.beginSession(time.Now())
	m.stopping = true

	s.stopRecord = entity.TaskRecord{TimeDuration: 12}
	s.stopErr = nil
	m, cmd := updateTimer(t, m, stopTaskResultMsg{err: errServerDown, abortPlan: true})
	if cmd == nil {
		t.Fatal("stop error: expected a retry, got nil")
	}
	msg := cmd()
	if _, quit := msg.(tea.QuitMsg); quit {
		t.Fatal("stop error inside the window quit the timer")
	}
	retry, ok := msg.(stopTaskResultMsg)
	if !ok {
		t.Fatalf("expected stopTaskResultMsg from the retry, got %T", msg)
	}
	if !retry.abortPlan {
		t.Fatal("retry lost abortPlan")
	}
	if got := m.getStatus(); got != "server unreachable, retrying" {
		t.Fatalf("status while retrying the stop = %q", got)
	}

	m, cmd = updateTimer(t, m, retry)
	if !m.exitState.shouldSave || !m.exitState.abortPlan {
		t.Fatalf("exitState = %+v, want shouldSave and abortPlan", m.exitState)
	}
	if m.task.TimeDone != 12 {
		t.Fatalf("TimeDone = %v, want 12", m.task.TimeDone)
	}
	if !m.stopFailingSince.IsZero() {
		t.Fatal("successful stop left stopFailingSince set")
	}
	if !isQuit(cmd) {
		t.Fatal("expected tea.Quit after a successful stop")
	}
}

func TestTeaTimerModel_StopErrorPastWindowExitsServerLost(t *testing.T) {
	s := swapRetrySeams(t)
	t0 := s.now
	m := newRetryTestModel()
	m.task.beginSession(time.Now())
	m.stopping = true

	m, _ = updateTimer(t, m, stopTaskResultMsg{err: errServerDown})
	s.now = t0.Add(31 * time.Second)
	m, cmd := updateTimer(t, m, stopTaskResultMsg{err: errServerDown})
	if !isQuit(cmd) {
		t.Fatal("expected tea.Quit once the stop window is over")
	}
	if !m.exitState.serverLost {
		t.Fatal("expected serverLost")
	}
}

func TestTeaTimerModel_PollIgnoredWhileStopping(t *testing.T) {
	swapRetrySeams(t)
	m := newRetryTestModel()
	m.stopping = true

	m, cmd := updateTimer(t, m, statusPollMsg{err: errServerDown})
	if cmd != nil {
		t.Fatal("expected no command for a poll while stopping")
	}
	if m.exitState != (exitState{}) {
		t.Fatalf("exitState = %+v, want zero", m.exitState)
	}
}

func TestTaskTimer_HandleExitServerLost(t *testing.T) {
	swapRetrySeams(t)
	var logBuf bytes.Buffer
	origLog := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(origLog) })

	m := newRetryTestModel()
	m.exitState = exitState{serverLost: true}

	var buf bytes.Buffer
	err := m.task.handleExit(&buf, m)
	if !errors.Is(err, ErrServerUnreachable) || !errors.Is(err, ErrTaskAborted) {
		t.Fatalf("handleExit error = %v, want ErrServerUnreachable wrapping ErrTaskAborted", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Server unreachable for 30s") || !strings.Contains(out, `"focus"`) {
		t.Fatalf("unexpected message: %q", out)
	}
	if strings.Contains(logBuf.String(), "timer cancelled without saving") {
		t.Fatalf("log has the cancel line: %q", logBuf.String())
	}
}
