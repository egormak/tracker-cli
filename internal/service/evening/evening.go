package evening

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	taskStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type model struct {
	category    string
	sprintTime  int
	focus       entity.EveningFocusResponse
	err         error
	skippedLast string
	action      string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.action = "quit"
			return m, tea.Quit
		case "enter":
			if m.focus.CurrentTask.TaskName != "" {
				m.action = "start"
				return m, tea.Quit
			}
		case "s", "tab":
			if m.focus.CurrentTask.TaskName != "" {
				m.skippedLast = m.focus.CurrentTask.TaskName
				nextFocus, err := api.SkipEveningTask(m.focus.CurrentTask.TaskName, m.category, m.sprintTime)
				if err != nil {
					m.err = err
				} else {
					m.focus = nextFocus
				}
			}
		case "1":
			m.sprintTime = 15
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err == nil {
				m.focus = nextFocus
			}
		case "2":
			m.sprintTime = 20
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err == nil {
				m.focus = nextFocus
			}
		case "3":
			m.sprintTime = 30
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err == nil {
				m.focus = nextFocus
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	b := strings.Builder{}
	b.WriteString(titleStyle.Render("🌙 РЕЖИМ ВЕЧЕРНЕГО ДОБОРА (Evening Catch-Up Mode)"))
	b.WriteString("\n")

	if m.skippedLast != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("ℹ️ Задача '%s' пропущена на сегодня", m.skippedLast)) + "\n\n")
	}

	if m.focus.CurrentTask.TaskName == "" {
		b.WriteString("🎉 Отличная работа! Все незавершенные задачи на эту неделю выполнены.\n")
	} else {
		taskInfo := fmt.Sprintf(
			"🎯 Рекомендуемая задача:  %s\n"+
				"📊 Сделано за неделю:     %d мин (план: %d мин)\n"+
				"⏱️ Целевой спринт:       %d мин (Отдых: %d мин)\n"+
				"📋 Кандидаты в очереди: %d",
			taskStyle.Render(m.focus.CurrentTask.TaskName),
			m.focus.CurrentTask.WeeklyDone,
			m.focus.CurrentTask.WeeklyTarget,
			m.sprintTime,
			m.focus.RestPool,
			len(m.focus.Candidates),
		)
		b.WriteString(boxStyle.Render(taskInfo) + "\n")
	}

	help := helpStyle.Render("[Enter] ▶️ Начать спринт  |  [s / Tab] ⏭️ Скип  |  [1: 15m | 2: 20m | 3: 30m]  |  [q] Выход")
	b.WriteString(help + "\n")

	return b.String()
}

func RunEveningSession(category string, sprintTime int, skipTask string) (selectedTask string, duration int, err error) {
	if sprintTime <= 0 {
		sprintTime = 20
	}

	if skipTask != "" {
		_, _ = api.SkipEveningTask(skipTask, category, sprintTime)
	}

	focus, err := api.GetEveningFocus(category, sprintTime)
	if err != nil {
		return "", 0, fmt.Errorf("failed to fetch evening focus: %w", err)
	}

	m := model{
		category:   category,
		sprintTime: sprintTime,
		focus:      focus,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", 0, err
	}

	res := finalModel.(model)
	if res.action == "start" && res.focus.CurrentTask.TaskName != "" {
		return res.focus.CurrentTask.TaskName, res.sprintTime, nil
	}

	return "", 0, nil
}
