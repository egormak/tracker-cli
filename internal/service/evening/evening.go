package evening

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/ui/theme"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1).
			Width(56)

	selectedCardStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Bold(true).
				Padding(0, 1).
				Width(56)

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type model struct {
	category      string
	sprintTime    int
	focus         entity.EveningFocusResponse
	err           error
	skippedLast   string
	action        string
	selectedIndex int
}

func (m model) Init() tea.Cmd { return nil }

func GetTopCandidates(focus entity.EveningFocusResponse) []entity.EveningFocusCandidate {
	candidates := focus.Candidates
	if len(candidates) == 0 && focus.CurrentTask.TaskName != "" {
		candidates = []entity.EveningFocusCandidate{focus.CurrentTask}
	}
	if len(candidates) > 3 {
		return candidates[:3]
	}
	return candidates
}

func (m model) getTopCandidates() []entity.EveningFocusCandidate {
	return GetTopCandidates(m.focus)
}

func (m *model) clampSelection() {
	top := m.getTopCandidates()
	if len(top) == 0 {
		m.selectedIndex = 0
		return
	}
	if m.selectedIndex >= len(top) {
		m.selectedIndex = len(top) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.action = "quit"
			return m, tea.Quit
		case "up", "k":
			top := m.getTopCandidates()
			if len(top) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(top) - 1
				}
			}
		case "down", "j":
			top := m.getTopCandidates()
			if len(top) > 0 {
				m.selectedIndex = (m.selectedIndex + 1) % len(top)
			}
		case "c", "C":
			top := m.getTopCandidates()
			if len(top) > 0 {
				m.action = "combo"
				return m, tea.Quit
			}
		case "enter":
			top := m.getTopCandidates()
			if len(top) > 0 && m.selectedIndex < len(top) {
				m.action = "start"
				return m, tea.Quit
			}
		case "s", "tab":
			top := m.getTopCandidates()
			if len(top) > 0 && m.selectedIndex < len(top) {
				taskToSkip := top[m.selectedIndex].TaskName
				m.skippedLast = taskToSkip
				nextFocus, err := api.SkipEveningTask(taskToSkip, m.category, m.sprintTime)
				if err != nil {
					m.err = err
				} else {
					m.err = nil
					m.focus = nextFocus
					m.clampSelection()
				}
			}
		case "1":
			m.sprintTime = 15
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err != nil {
				m.err = err
			} else {
				m.err = nil
				m.focus = nextFocus
				m.clampSelection()
			}
		case "2":
			m.sprintTime = 20
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err != nil {
				m.err = err
			} else {
				m.err = nil
				m.focus = nextFocus
				m.clampSelection()
			}
		case "3":
			m.sprintTime = 30
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err != nil {
				m.err = err
			} else {
				m.err = nil
				m.focus = nextFocus
				m.clampSelection()
			}
		case "t", "T":
			switch m.sprintTime {
			case 15:
				m.sprintTime = 20
			case 20:
				m.sprintTime = 30
			case 30:
				m.sprintTime = 15
			default:
				m.sprintTime = 20
			}
			nextFocus, err := api.GetEveningFocus(m.category, m.sprintTime)
			if err != nil {
				m.err = err
			} else {
				m.err = nil
				m.focus = nextFocus
				m.clampSelection()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	b := strings.Builder{}
	b.WriteString(titleStyle.Render("🌙 РЕЖИМ ВЕЧЕРНЕГО ДОБОРА (Evening Catch-Up 2.0)"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("⚠️ %v", m.err)) + "\n\n")
	}

	if m.skippedLast != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("ℹ️ Task '%s' skipped for tonight", m.skippedLast)) + "\n\n")
	}

	top := m.getTopCandidates()
	if len(top) == 0 {
		b.WriteString("🎉 All weekly tasks completed.\n\n")
		b.WriteString(helpStyle.Render("[q] Quit") + "\n")
		return b.String()
	}

	for i, cand := range top {
		isSelected := (i == m.selectedIndex)

		cursor := "  "
		style := cardStyle
		if isSelected {
			cursor = cursorStyle.Render("❯ ")
			style = selectedCardStyle
		}

		rankText := fmt.Sprintf("[%d]", i+1)
		rank := lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(rankText)
		if isSelected {
			rank = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(rankText)
		}

		taskName := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(cand.TaskName)
		if isSelected {
			taskName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(cand.TaskName)
		}

		var roleTag string
		if cand.Role != "" {
			roleTag = " " + theme.RoleBadge(cand.Role)
		}

		ratio := 0.0
		if cand.WeeklyTarget > 0 {
			ratio = float64(cand.WeeklyDone) / float64(cand.WeeklyTarget)
		}

		barColor := theme.RoleColor(cand.Role)
		if cand.Role == "" {
			barColor = lipgloss.Color("86")
		}
		bar := fmt.Sprintf("[%s]", theme.RenderProgressBar(10, ratio, barColor))

		gap := cand.WeeklyGap
		if gap == 0 && cand.WeeklyTarget > cand.WeeklyDone {
			gap = cand.WeeklyTarget - cand.WeeklyDone
		}

		line1 := fmt.Sprintf("%s%s %s%s", cursor, rank, taskName, roleTag)
		line2 := fmt.Sprintf("   %s %d/%dm (deficit: %dm)", bar, cand.WeeklyDone, cand.WeeklyTarget, gap)

		b.WriteString(style.Render(line1+"\n"+line2) + "\n")
	}

	// Duration selector: [1] 15m  [2] 20m  [3] 30m
	durations := []struct {
		key string
		val int
	}{
		{"1", 15},
		{"2", 20},
		{"3", 30},
	}
	var durItems []string
	for _, d := range durations {
		txt := fmt.Sprintf("[%s] %dm", d.key, d.val)
		if m.sprintTime == d.val {
			durItems = append(durItems, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(txt))
		} else {
			durItems = append(durItems, lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(txt))
		}
	}
	durLine := fmt.Sprintf("⏱️ Duration: %s (or toggle with 't')", strings.Join(durItems, "  "))
	if m.focus.RestPool > 0 {
		durLine += fmt.Sprintf("  |  Rest Pool: %d min", m.focus.RestPool)
	}
	b.WriteString("\n" + durLine + "\n")

	// Keymap instructions
	help := helpStyle.Render("↑/↓ or j/k: Select task | Enter: Start | s: Skip | c: Launch Combo 3x10m | q: Quit")
	b.WriteString(help + "\n")

	return b.String()
}

func RunEveningSession(category string, sprintTime int, skipTask string) (selectedTask string, duration int, isCombo bool, comboCandidates []entity.EveningFocusCandidate, err error) {
	if sprintTime <= 0 {
		sprintTime = 20
	}

	if skipTask != "" {
		_, _ = api.SkipEveningTask(skipTask, category, sprintTime)
	}

	focus, err := api.GetEveningFocus(category, sprintTime)
	if err != nil {
		return "", 0, false, nil, fmt.Errorf("failed to fetch evening focus: %w", err)
	}

	m := model{
		category:      category,
		sprintTime:    sprintTime,
		focus:         focus,
		selectedIndex: 0,
	}
	m.clampSelection()

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", 0, false, nil, err
	}

	res := finalModel.(model)
	top := res.getTopCandidates()

	if res.action == "combo" {
		return "", res.sprintTime, true, top, nil
	}

	if res.action == "start" && len(top) > 0 && res.selectedIndex < len(top) {
		return top[res.selectedIndex].TaskName, res.sprintTime, false, top, nil
	}

	return "", 0, false, nil, nil
}
