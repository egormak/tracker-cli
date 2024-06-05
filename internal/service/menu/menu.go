package menu

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
	"tracker_cli/config"

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

	type taskStat struct {
		Name     string `json:"name"`
		Role     string `json:"role"`
		Priority int    `json:"priority"`
		Duration int    `json:"time_duration"`
		Done     int    `json:"time_done"`
	}

	var rows []table.Row

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/tasklist"), nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal(fmt.Errorf("request error, status code: %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	var tasksInfo []taskStat
	err = json.NewDecoder(resp.Body).Decode(&tasksInfo)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to decode response: %w", err))
	}

	sort.Slice(tasksInfo, func(i, j int) bool { return tasksInfo[i].Priority > tasksInfo[j].Priority })

	for _, task := range tasksInfo {
		rows = append(rows, table.Row{task.Name, task.Role, strconv.Itoa(task.Priority), strconv.Itoa(task.Duration), strconv.Itoa(task.Done), strconv.Itoa(task.Duration - task.Done)})
	}

	return rows
}
