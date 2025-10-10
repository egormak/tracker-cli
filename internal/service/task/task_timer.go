package task

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	bubblestimer "github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/procent"
	"tracker_cli/internal/service/rest"
	"tracker_cli/internal/service/statistic"
	"tracker_cli/internal/service/telegram"
	"tracker_cli/internal/service/timer"
)

type timerKeymap struct {
	pause key.Binding
	stop  key.Binding
	quit  key.Binding
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
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit without saving"),
		),
		help: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "toggle help"),
		),
	}
}

type teaTimerModel struct {
	timer      bubblestimer.Model
	keymap     timerKeymap
	duration   time.Duration
	elapsed    time.Duration
	quit       bool
	task       *TaskTimer
	sendResult bool
	completed  bool
	showHelp   bool
}

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
		m.sendResult = true
		m.completed = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.syncElapsed()
			m.sendResult = true
			m.quit = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.pause):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymap.stop):
			if !m.task.started() {
				return m, nil
			}
			m.syncElapsed()
			m.sendResult = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.help):
			m.showHelp = !m.showHelp
			return m, nil
		}
	}
	return m, nil
}

func (m teaTimerModel) View() string {
	status := "paused"
	if m.timer.Running() {
		status = "running"
	} else if m.timer.Timedout() {
		status = "completed"
	} else if m.quit {
		status = "quitting"
	}

	remaining := m.duration - m.elapsed
	if remaining < 0 {
		remaining = 0
	}

	view := fmt.Sprintf(
		"Task: %s\nStatus: %s\nElapsed: %s\nRemaining: %s\n\n",
		m.task.Name,
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
			fmt.Sprintf("  %-10s %s", "q / ctrl+c", "quit and save progress"),
			fmt.Sprintf("  %-10s %s", "h", "toggle this help view"),
		}
		return strings.Join(lines, "\n")
	}

	return "Controls: p pause/resume • enter stop-save • q quit-save • h help"
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
	program := tea.NewProgram(model, tea.WithAltScreen())

	result, err := program.Run()
	if err != nil {
		return err
	}

	timerResult, ok := result.(teaTimerModel)
	if !ok {
		return fmt.Errorf("invalid timer result")
	}

	if !timerResult.sendResult {
		slog.Info("timer cancelled", "task", t.Name)
		return nil
	}

	t.finalizeSession(timerResult.elapsed, timerResult.completed)
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

func (t *TaskTimer) finalizeSession(elapsed time.Duration, completed bool) {
	if !t.started() {
		slog.Info("task timer exited before start", "task", t.Name)
		return
	}

	t.TimeEnd = t.TimeBegin.Add(elapsed)
	t.TimeDone = elapsed.Minutes()

	if completed {
		timer.TimeDurationDel(t.TimeDuration)
		if err := procent.ChangeGroupPlanPercent(); err != nil {
			slog.Error("failed to notify percent change", "error", err)
		}
	}

	minutesLogged := int(t.TimeDone)
	if minutesLogged > 0 {
		api.AddTaskRecord(t.Name, minutesLogged)
	}

	statistic.StatisticTaskShow(t.Name)
	statistic.StatisticFullShow()
	rest.RestShow()

	telegram.TelegramStopSend(t.Name, t.MsgID, minutesLogged, t.TimeEnd.Format("2 January 2006 15:04"))
}
