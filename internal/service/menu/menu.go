package menu

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"tracker_cli/internal/repository/api"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table  table.Model
	choose bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.choose = false
			return m, tea.Quit
		case "enter":
			m.choose = true
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func RunMenu() string {
	columns := []table.Column{
		{Title: "Name", Width: 15},
		{Title: "Role", Width: 7},
		{Title: "Priority", Width: 10},
		{Title: "Duration", Width: 10},
		{Title: "Done", Width: 5},
		{Title: "% Done", Width: 7},
		{Title: "Left", Width: 5},
	}

	rows := GetRows()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(9),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t, false}

	p := tea.NewProgram(m)

	r, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if r == nil {
		return ""
	}

	if r, ok := r.(model); ok {
		if r.choose {
			return r.table.SelectedRow()[0]
		} else {
			return ""
		}
	}

	return ""
}

func GetRows() []table.Row {
	var rows []table.Row

	tasksInfo, err := api.GetTaskList()
	if err != nil {
		fmt.Printf("Error fetching task list: %v\n", err)
		return rows
	}

	sort.Slice(tasksInfo, func(i, j int) bool { return tasksInfo[i].Priority > tasksInfo[j].Priority })

	for _, task := range tasksInfo {
		percentDone := "0%"
		if task.TimeDuration > 0 {
			value := float64(task.TimeDone) / float64(task.TimeDuration) * 100
			percentDone = fmt.Sprintf("%.0f%%", value)
		}
		rows = append(rows, table.Row{
			task.Name,
			task.Role,
			strconv.Itoa(task.Priority),
			strconv.Itoa(task.TimeDuration),
			strconv.Itoa(task.TimeDone),
			percentDone,
			strconv.Itoa(task.TimeDuration - task.TimeDone),
		})
	}

	return rows
}
