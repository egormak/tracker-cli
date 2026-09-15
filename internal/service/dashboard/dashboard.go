package dashboard

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/pkg/restutil"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/task"
	"tracker_cli/internal/ui/theme"
)

type activeTab int

const (
	tabTimer activeTab = iota
	tabPlan
	tabAnalytics
	tabEvening
)

type DashboardModel struct {
	activeTab     activeTab
	runningTask   entity.RunningTask
	taskList      []entity.TaskList
	rolloverTasks []entity.RolloverTask
	eveningFocus  entity.EveningFocusResponse
	restUnits     int
	scheduledTime int
	completedTime int
	tableModel    table.Model
	statusMsg     string
	errMsg        string
	width         int
	height        int
	loading       bool
	quitting      bool
	taskToRun     string
	sourceDay     string
}

type dataLoadedMsg struct {
	runningTask   entity.RunningTask
	taskList      []entity.TaskList
	rolloverTasks []entity.RolloverTask
	eveningFocus  entity.EveningFocusResponse
	restUnits     int
	scheduledTime int
	completedTime int
	err           error
}

type tickMsg time.Time

func fetchDashboardDataCmd() tea.Cmd {
	return func() tea.Msg {
		msg := dataLoadedMsg{}

		// 1. Running Task
		rt, err := api.GetRunningTaskStatus("")
		if err == nil {
			msg.runningTask = rt
		}

		// 2. Task List
		tl, err := api.GetTaskList()
		if err == nil {
			msg.taskList = tl
		}

		// 3. Rollover Tasks
		ro, err := api.GetRolloverTasks()
		if err == nil {
			msg.rolloverTasks = ro
		}

		// 4. Rest Units
		ru, err := api.GetRestTime()
		if err == nil {
			msg.restUnits = ru
		}

		// 5. Scheduled & Completed Time
		st, _ := api.GetScheduledTimeToday()
		msg.scheduledTime = st
		ct, _ := api.GetCompletionTimeDoneToday()
		msg.completedTime = ct

		// 6. Evening Focus
		ef, _ := api.GetEveningFocus("", 20)
		msg.eveningFocus = ef

		return msg
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewDashboardModel() DashboardModel {
	columns := []table.Column{
		{Title: "Task", Width: 20},
		{Title: "Role", Width: 10},
		{Title: "Target", Width: 8},
		{Title: "Done", Width: 8},
		{Title: "Left", Width: 8},
		{Title: "Progress", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.ColorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.ColorText)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4F46E5")).
		Bold(true)
	t.SetStyles(s)

	return DashboardModel{
		activeTab:  tabTimer,
		tableModel: t,
		loading:    true,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(fetchDashboardDataCmd(), tickCmd())
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.runningTask = msg.runningTask
		m.taskList = msg.taskList
		m.rolloverTasks = msg.rolloverTasks
		m.restUnits = msg.restUnits
		m.scheduledTime = msg.scheduledTime
		m.completedTime = msg.completedTime
		m.eveningFocus = msg.eveningFocus

		// Build table rows
		var rows []table.Row
		sort.Slice(m.taskList, func(i, j int) bool { return m.taskList[i].Priority > m.taskList[j].Priority })

		for _, t := range m.taskList {
			left := t.TimeDuration - t.TimeDone
			if left < 0 {
				left = 0
			}
			progress := "0%"
			if t.TimeDuration > 0 {
				progress = fmt.Sprintf("%.0f%%", float64(t.TimeDone)/float64(t.TimeDuration)*100)
			}
			rows = append(rows, table.Row{
				t.Name,
				strings.ToUpper(t.Role),
				strconv.Itoa(t.TimeDuration),
				strconv.Itoa(t.TimeDone),
				strconv.Itoa(left),
				progress,
			})
		}
		m.tableModel.SetRows(rows)
		return m, nil

	case tickMsg:
		return m, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
			return m, nil

		case "shift+tab":
			m.activeTab = (m.activeTab + 3) % 4
			return m, nil

		case "1":
			m.activeTab = tabTimer
			return m, nil
		case "2":
			m.activeTab = tabPlan
			return m, nil
		case "3":
			m.activeTab = tabAnalytics
			return m, nil
		case "4":
			m.activeTab = tabEvening
			return m, nil

		case "r":
			m.loading = true
			m.statusMsg = "Refreshing data..."
			return m, fetchDashboardDataCmd()

		case " ":
			// Toggle pause on running task if in timer tab
			if m.runningTask.TaskName != "" {
				if m.runningTask.IsRunning {
					_, _ = api.PauseRunningTask(m.runningTask.TaskName)
				} else {
					_, _ = api.ResumeRunningTask(m.runningTask.TaskName)
				}
				return m, fetchDashboardDataCmd()
			}

		case "+", "=", "]":
			if m.runningTask.TaskName != "" {
				_, _ = api.AdjustRunningTask(m.runningTask.TaskName, 5)
				return m, fetchDashboardDataCmd()
			}

		case "-", "_", "[":
			if m.runningTask.TaskName != "" {
				_, _ = api.AdjustRunningTask(m.runningTask.TaskName, -5)
				return m, fetchDashboardDataCmd()
			}

		case "enter":
			if m.activeTab == tabPlan {
				selected := m.tableModel.SelectedRow()
				if len(selected) > 0 {
					m.taskToRun = selected[0]
					return m, tea.Quit
				}
			} else if m.activeTab == tabEvening {
				if m.eveningFocus.CurrentTask.TaskName != "" {
					m.taskToRun = m.eveningFocus.CurrentTask.TaskName
					return m, tea.Quit
				}
			} else if m.activeTab == tabTimer && m.runningTask.TaskName != "" {
				_, _ = api.StopRunningTask(m.runningTask.TaskName)
				m.statusMsg = fmt.Sprintf("Stopped task %s", m.runningTask.TaskName)
				return m, fetchDashboardDataCmd()
			}

		case "s":
			if m.activeTab == tabEvening && m.eveningFocus.CurrentTask.TaskName != "" {
				ef, err := api.SkipEveningTask(m.eveningFocus.CurrentTask.TaskName, "", 20)
				if err == nil {
					m.eveningFocus = ef
					m.statusMsg = "Skipped evening task"
				}
				return m, nil
			}
		}
	}

	if m.activeTab == tabPlan {
		var cmd tea.Cmd
		m.tableModel, cmd = m.tableModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	header := m.renderHeader()
	var body string

	switch m.activeTab {
	case tabTimer:
		body = m.renderTimerTab()
	case tabPlan:
		body = m.renderPlanTab()
	case tabAnalytics:
		body = m.renderAnalyticsTab()
	case tabEvening:
		body = m.renderEveningTab()
	}

	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m DashboardModel) renderHeader() string {
	title := theme.TitleStyle.Render("⚡ TRACKER TIMEFLOW TUI")

	tabs := []string{
		"1. ⏱ Timer",
		"2. 📋 Today Plan",
		"3. 📊 Analytics",
		"4. 🌙 Evening Mode",
	}

	var renderedTabs []string
	for i, t := range tabs {
		if activeTab(i) == m.activeTab {
			renderedTabs = append(renderedTabs, theme.ActiveTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, theme.TabStyle.Render(t))
		}
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	return lipgloss.JoinVertical(lipgloss.Left, title, tabRow, "")
}

func (m DashboardModel) renderTimerTab() string {
	if m.runningTask.TaskName == "" {
		return theme.CardStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				theme.TitleStyle.Render("No Active Timer"),
				theme.SubTitleStyle.Render("Press [2] to pick a task from Today's Plan, or run 'tracker task -n <name>'"),
			),
		)
	}

	roleBadge := theme.RoleBadge(m.runningTask.Role)
	statusStr := "● RUNNING"
	statusColor := theme.ColorSuccess
	if !m.runningTask.IsRunning {
		statusStr = "⏸ PAUSED"
		statusColor = theme.ColorWarning
	}

	statusBadge := lipgloss.NewStyle().Bold(true).Foreground(statusColor).Render(statusStr)

	elapsed := time.Duration(m.runningTask.Accumulated) * time.Minute
	if m.runningTask.IsRunning && !m.runningTask.StartTime.IsZero() {
		elapsed += time.Since(m.runningTask.StartTime)
	}

	target := time.Duration(m.runningTask.TargetDuration) * time.Minute
	remaining := target - elapsed
	if remaining < 0 {
		remaining = 0
	}

	ratio := 0.0
	if target > 0 {
		ratio = float64(elapsed) / float64(target)
	}

	progressBar := theme.RenderProgressBar(36, ratio, theme.RoleColor(m.runningTask.Role))

	content := fmt.Sprintf(
		"%s  %s  %s\n\nElapsed:   %02dm %02ds\nTarget:    %02dm 00s\nRemaining: %02dm %02ds\n\n[%s] %3.0f%%\n\nControls: [Space] Pause/Resume • [+/-] Adjust 5m • [Enter] Stop & Save",
		roleBadge,
		lipgloss.NewStyle().Bold(true).Foreground(theme.ColorText).Render(m.runningTask.TaskName),
		statusBadge,
		int(elapsed.Minutes()), int(elapsed.Seconds())%60,
		int(target.Minutes()),
		int(remaining.Minutes()), int(remaining.Seconds())%60,
		progressBar,
		ratio*100,
	)

	return theme.ActiveCardStyle.Render(content)
}

func (m DashboardModel) renderPlanTab() string {
	tableStr := m.tableModel.View()
	instructions := theme.SubTitleStyle.Render("↑/↓ Navigate • [Enter] Launch Selected Task • [r] Refresh")
	return theme.CardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, tableStr, "", instructions))
}

func (m DashboardModel) renderAnalyticsTab() string {
	scheduled := m.scheduledTime
	completed := m.completedTime
	percent := 0.0
	if scheduled > 0 {
		percent = float64(completed) / float64(scheduled) * 100
	}

	ratio := 0.0
	if scheduled > 0 {
		ratio = float64(completed) / float64(scheduled)
	}

	progressBar := theme.RenderProgressBar(40, ratio, theme.ColorWork)
	restMinutes := restutil.MinutesFromUnits(m.restUnits)

	stats := fmt.Sprintf(
		"Daily Target: %d min\nCompleted:    %d min (%1.0f%%)\n\n[%s]\n\nRest Pool Balance: %.1f min (Time to relax or exercise!)\n",
		scheduled, completed, percent, progressBar, restMinutes,
	)

	return theme.CardStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			theme.TitleStyle.Render("📊 DAILY PROGRESS & REST BALANCE"),
			stats,
		),
	)
}

func (m DashboardModel) renderEveningTab() string {
	if m.eveningFocus.CurrentTask.TaskName == "" {
		return theme.CardStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				theme.TitleStyle.Render("🌙 Evening Catch-Up Mode"),
				theme.SubTitleStyle.Render("All weekly tasks are completed or no deficit tasks found! 🎉"),
			),
		)
	}

	t := m.eveningFocus.CurrentTask
	badge := theme.RoleBadge(t.Role)
	title := fmt.Sprintf("Next Priority Task: %s", t.TaskName)

	info := fmt.Sprintf(
		"%s %s\nWeekly Gap: %d min | Sprint: %d min\n\nControls: [Enter] Start Sprint • [s] Skip Task • [r] Refresh",
		badge,
		lipgloss.NewStyle().Bold(true).Foreground(theme.ColorText).Render(title),
		t.WeeklyGap,
		m.eveningFocus.SprintTime,
	)

	return theme.ActiveCardStyle.Render(info)
}

func (m DashboardModel) renderFooter() string {
	status := m.statusMsg
	if m.errMsg != "" {
		status = lipgloss.NewStyle().Foreground(theme.ColorDanger).Render("Error: " + m.errMsg)
	}
	help := theme.SubTitleStyle.Render("[Tab/1-4] Switch View • [r] Refresh • [q] Quit")
	if status != "" {
		return lipgloss.JoinVertical(lipgloss.Left, "", lipgloss.NewStyle().Foreground(theme.ColorWarning).Render(status), help)
	}
	return lipgloss.JoinVertical(lipgloss.Left, "", help)
}

func RunDashboard() error {
	m := NewDashboardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	res, err := p.Run()
	if err != nil {
		return fmt.Errorf("dashboard run error: %w", err)
	}

	if finalModel, ok := res.(DashboardModel); ok && finalModel.taskToRun != "" {
		taskApp, err := task.CreateTaskTimer(finalModel.taskToRun, 0, 100)
		if err != nil {
			return err
		}
		if finalModel.sourceDay != "" {
			taskApp.SourceDay = finalModel.sourceDay
		}
		return taskApp.Run()
	}

	return nil
}
