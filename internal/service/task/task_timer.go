package task

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/pkg/restutil"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/procent"
	"tracker_cli/internal/service/rest"
	"tracker_cli/internal/service/statistic"
	"tracker_cli/internal/service/telegram"
)

const (
	// timeFormat is the display format for task completion time in Telegram
	timeFormat = "2 January 2006 15:04"
)

type timerKeymap struct {
	pause key.Binding
	stop  key.Binding
	quit  key.Binding
	abort key.Binding
	help  key.Binding
}

func newTimerKeymap() timerKeymap {
	return timerKeymap{
		pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause/resume"),
		),
		stop: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "stop & save"),
		),
		quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit & save"),
		),
		abort: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit & save, abort plan"),
		),
		help: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "toggle help"),
		),
	}
}

type teaTimerModel struct {
	keymap      timerKeymap
	duration    time.Duration
	elapsed     time.Duration
	accumulated time.Duration
	startTime   time.Time
	isRunning   bool
	task        *TaskTimer
	exitState   exitState
	showHelp    bool
}

type statusPollMsg struct {
	task entity.RunningTask
	err  error
}

type smoothTickMsg time.Time

type togglePauseResultMsg struct {
	task entity.RunningTask
	err  error
}

type stopTaskResultMsg struct {
	record    entity.TaskRecord
	abortPlan bool
	err       error
}

type exitState struct {
	shouldSave bool
	abortPlan  bool
	completed  bool
}

type interruptMsg struct{}

func pollStatusCmd() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
		status, err := api.GetRunningTaskStatus()
		return statusPollMsg{task: status, err: err}
	})
}

func smoothTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return smoothTickMsg(t)
	})
}

func togglePauseCmd(currentlyRunning bool) tea.Cmd {
	return func() tea.Msg {
		var task entity.RunningTask
		var err error
		if currentlyRunning {
			task, err = api.PauseRunningTask()
		} else {
			task, err = api.ResumeRunningTask()
		}
		return togglePauseResultMsg{task: task, err: err}
	}
}

func stopTaskCmd(abortPlan bool) tea.Cmd {
	return func() tea.Msg {
		record, err := api.StopRunningTask()
		return stopTaskResultMsg{record: record, abortPlan: abortPlan, err: err}
	}
}

func newTeaTimerModel(task *TaskTimer, runningTask entity.RunningTask) teaTimerModel {
	duration := time.Duration(task.TimeDuration) * time.Minute
	return teaTimerModel{
		keymap:      newTimerKeymap(),
		duration:    duration,
		task:        task,
		isRunning:   runningTask.IsRunning,
		accumulated: time.Duration(runningTask.Accumulated) * time.Minute,
		startTime:   runningTask.StartTime,
	}
}

func (m teaTimerModel) Init() tea.Cmd {
	m.task.beginSession(m.startTime)
	return tea.Batch(pollStatusCmd(), smoothTickCmd())
}

func (m teaTimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case smoothTickMsg:
		m.updateElapsed()
		return m, smoothTickCmd()

	case statusPollMsg:
		if msg.err != nil {
			slog.Error("Failed to fetch running task status from server", "error", msg.err)
			return m, tea.Quit
		}
		if msg.task.TaskName == "" || msg.task.TaskName != m.task.Name {
			// Task was stopped or completed from another service!
			m.exitState = exitState{shouldSave: false, completed: true, abortPlan: false}
			return m, tea.Quit
		}
		m.isRunning = msg.task.IsRunning
		m.accumulated = time.Duration(msg.task.Accumulated) * time.Minute
		m.startTime = msg.task.StartTime
		m.updateElapsed()
		return m, pollStatusCmd()

	case togglePauseResultMsg:
		if msg.err != nil {
			slog.Error("Failed to toggle pause on server", "error", msg.err)
			return m, nil
		}
		m.isRunning = msg.task.IsRunning
		m.accumulated = time.Duration(msg.task.Accumulated) * time.Minute
		m.startTime = msg.task.StartTime
		m.updateElapsed()
		return m, nil

	case stopTaskResultMsg:
		if msg.err != nil {
			slog.Error("Failed to stop running task on server", "error", msg.err)
			return m, tea.Quit
		}
		m.task.TimeEnd = time.Now()
		m.task.TimeDone = float64(msg.record.TimeDuration)
		m.exitState = exitState{shouldSave: true, completed: m.elapsed >= m.duration, abortPlan: msg.abortPlan}
		return m, tea.Quit

	case interruptMsg:
		return m, stopTaskCmd(true)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, stopTaskCmd(false)
		case key.Matches(msg, m.keymap.abort):
			return m, stopTaskCmd(true)
		case key.Matches(msg, m.keymap.pause):
			return m, togglePauseCmd(m.isRunning)
		case key.Matches(msg, m.keymap.stop):
			if !m.task.started() {
				return m, nil
			}
			return m, stopTaskCmd(false)
		case key.Matches(msg, m.keymap.help):
			m.showHelp = !m.showHelp
			return m, nil
		}
	}
	return m, nil
}

func (m teaTimerModel) View() string {
	status := m.getStatus()
	remaining := m.getRemainingTime()

	taskInfo := fmt.Sprintf("Task: %s", m.task.Name)
	if m.task.SourceDay != "" {
		taskInfo = fmt.Sprintf("%s (rollover from %s)", taskInfo, m.task.SourceDay)
	}

	view := fmt.Sprintf(
		"%s\nStatus: %s\nElapsed: %s\nRemaining: %s\n\n",
		taskInfo,
		status,
		formatDuration(m.elapsed),
		formatDuration(remaining),
	)

	view += m.instructions()

	return view
}

func (m teaTimerModel) instructions() string {
	if m.showHelp {
		lines := []string{
			"Controls:",
			fmt.Sprintf("  %-10s %s", "p", "pause or resume timer"),
			fmt.Sprintf("  %-10s %s", "enter", "stop timer and save result"),
			fmt.Sprintf("  %-10s %s", "q", "quit and save progress"),
			fmt.Sprintf("  %-10s %s", "ctrl+c", "quit, save, and abort plan"),
			fmt.Sprintf("  %-10s %s", "h", "toggle this help view"),
		}
		return strings.Join(lines, "\n")
	}

	return "Controls: p pause/resume • enter stop-save • q quit-save • ctrl+c abort-plan • h help"
}

// getStatus returns the current timer status as a string
func (m teaTimerModel) getStatus() string {
	if m.exitState.shouldSave {
		return "quitting"
	}
	if m.isRunning {
		return "running"
	}
	return "paused"
}

// getRemainingTime calculates and clamps the remaining time
func (m teaTimerModel) getRemainingTime() time.Duration {
	remaining := m.duration - m.elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (m *teaTimerModel) updateElapsed() {
	if m.isRunning && !m.startTime.IsZero() {
		m.elapsed = m.accumulated + time.Since(m.startTime)
	} else {
		m.elapsed = m.accumulated
	}
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02dm %02ds", minutes, seconds)
}

func (t *TaskTimer) Run() error {
	// Call server to start the running task
	runningTask, err := api.StartRunningTask(t.Name, t.Role, t.TimeDuration, t.SourceDay)
	if err != nil {
		return fmt.Errorf("failed to start task on server: %w", err)
	}

	model := newTeaTimerModel(t, runningTask)

	// Create program with signal catching disabled so Bubble Tea handles ctrl+c naturally
	program := tea.NewProgram(model, tea.WithoutSignalHandler())

	// Set up our own signal handler that works properly with Bubble Tea
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	// Launch signal forwarder in background
	stopForwarder := make(chan struct{})
	go func() {
		for {
			select {
			case <-sigCh:
				program.Send(interruptMsg{})
			case <-stopForwarder:
				return
			}
		}
	}()

	result, err := program.Run()
	close(stopForwarder) // Signal goroutine to exit

	if err != nil {
		return fmt.Errorf("running timer UI: %w", err)
	}

	timerResult, ok := result.(teaTimerModel)
	if !ok {
		return fmt.Errorf("invalid timer result: expected teaTimerModel, got %T", result)
	}

	// Handle exit state
	if !timerResult.exitState.shouldSave {
		slog.Info("timer cancelled without saving", "task", t.Name)
		return nil
	}

	t.finalizeSession(timerResult.elapsed, timerResult.exitState.completed)

	if timerResult.exitState.abortPlan {
		return ErrTaskAborted
	}
	return nil
}

func (t *TaskTimer) beginSession(startTime time.Time) {
	if t.TimeBegin.IsZero() {
		t.TimeBegin = startTime
		slog.Info("timer started on server", "task", t.Name, "duration_minutes", t.TimeDuration)
	}
}

func (t *TaskTimer) started() bool {
	return !t.TimeBegin.IsZero()
}

func (t *TaskTimer) SetRestLimitActive(active bool) {
	t.restLimitActive = active
}

func (t *TaskTimer) finalizeSession(elapsed time.Duration, completed bool) {
	if !t.started() {
		slog.Info("task timer exited before start", "task", t.Name)
		return
	}

	t.TimeEnd = t.TimeBegin.Add(elapsed)
	t.TimeDone = elapsed.Minutes()

	if completed {
		notifyMessage := ""
		if message, err := procent.ChangeGroupPlanPercent(); err != nil {
			slog.Error("failed to notify percent change", "error", err)
			notifyMessage = ""
		} else {
			notifyMessage = message
		}

		if !t.restLimitActive {
			if restUnits, err := api.GetRestTime(); err != nil {
				slog.Error("failed to fetch rest balance for notification", "error", err)
			} else {
				restMinutes := restutil.MinutesFromUnits(restUnits)
				if restMinutes > 0 {
					restStatement := fmt.Sprintf("Rest balance %.1f minutes. Time to rest or do some exercise.", restMinutes)
					if notifyMessage != "" {
						notifyMessage = fmt.Sprintf("%s %s", notifyMessage, restStatement)
					} else if t.Percent > 0 {
						notifyMessage = fmt.Sprintf("Completed planned task '%s' (%d%%). %s", t.Name, t.Percent, restStatement)
					} else {
						notifyMessage = fmt.Sprintf("Completed task '%s'. %s", t.Name, restStatement)
					}
				}
			}
		}

		if msg := strings.TrimSpace(notifyMessage); msg != "" {
			telegram.TelegramMessageSend(msg)
		}
	}

	statistic.StatisticTaskShow(t.Name)
	statistic.StatisticFullShow()
	rest.RestShow()
}
