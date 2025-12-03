package task

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	bubblestimer "github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"tracker_cli/internal/pkg/restutil"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/procent"
	"tracker_cli/internal/service/rest"
	"tracker_cli/internal/service/statistic"
	"tracker_cli/internal/service/telegram"
	"tracker_cli/internal/service/timer"
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
	timer     bubblestimer.Model
	keymap    timerKeymap
	duration  time.Duration
	elapsed   time.Duration
	task      *TaskTimer
	exitState exitState
	showHelp  bool
}

// exitState represents the state when the timer exits
type exitState struct {
	shouldSave bool
	abortPlan  bool
	completed  bool
}

type interruptMsg struct{}

func newTeaTimerModel(task *TaskTimer) teaTimerModel {
	duration := time.Duration(task.TimeDuration) * time.Minute
	timerModel := bubblestimer.NewWithInterval(duration, time.Second)

	return teaTimerModel{
		timer:    timerModel,
		keymap:   newTimerKeymap(),
		duration: duration,
		task:     task,
	}
}

func (m teaTimerModel) Init() tea.Cmd {
	m.syncElapsed()
	m.task.beginSession()
	return m.timer.Init()
}

func (m teaTimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bubblestimer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.syncElapsed()
		return m, cmd
	case bubblestimer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		if m.timer.Running() {
			m.task.beginSession()
		}
		m.syncElapsed()
		return m, cmd
	case bubblestimer.TimeoutMsg:
		m.syncElapsed()
		m.exitState = exitState{shouldSave: true, completed: true, abortPlan: false}
		return m, tea.Quit
	case interruptMsg:
		m.syncElapsed()
		m.exitState = exitState{shouldSave: true, completed: false, abortPlan: true}
		return m, tea.Quit
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.syncElapsed()
			m.exitState = exitState{shouldSave: true, completed: false, abortPlan: false}
			return m, tea.Quit
		case key.Matches(msg, m.keymap.abort):
			m.syncElapsed()
			m.exitState = exitState{shouldSave: true, completed: false, abortPlan: true}
			return m, tea.Quit
		case key.Matches(msg, m.keymap.pause):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymap.stop):
			if !m.task.started() {
				return m, nil
			}
			m.syncElapsed()
			m.exitState = exitState{shouldSave: true, completed: false, abortPlan: false}
			return m, tea.Quit
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

// getStatus returns the current timer status as a string
func (m teaTimerModel) getStatus() string {
	if m.timer.Timedout() {
		return "completed"
	}
	if m.exitState.shouldSave {
		return "quitting"
	}
	if m.timer.Running() {
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

func (m *teaTimerModel) syncElapsed() {
	remaining := m.timer.Timeout
	if remaining < 0 {
		remaining = 0
	}
	elapsed := m.duration - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	m.elapsed = elapsed
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02dm %02ds", minutes, seconds)
}

func (t *TaskTimer) Run() error {
	model := newTeaTimerModel(t)

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

func (t *TaskTimer) beginSession() {
	if t.TimeBegin.IsZero() {
		t.TimeBegin = time.Now()
		t.MsgID = telegram.TelegramStartSend(t.Name)
		slog.Info("timer started", "task", t.Name, "duration_minutes", t.TimeDuration)
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
		timer.TimeDurationDel(t.TimeDuration)
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

	minutesLogged := int(t.TimeDone)
	if minutesLogged > 0 {
		api.AddTaskRecord(t.Name, minutesLogged, t.SourceDay)
	}

	statistic.StatisticTaskShow(t.Name)
	statistic.StatisticFullShow()
	rest.RestShow()

	telegram.TelegramStopSend(t.Name, t.MsgID, minutesLogged, t.TimeEnd.Format(timeFormat))
}
