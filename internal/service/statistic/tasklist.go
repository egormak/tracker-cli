package statistic

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func ShowTaskNameList() {
	taskList, err := api.GetTaskList()
	if err != nil {
		slog.Error("failed to fetch task list", "error", err)
		os.Exit(1)
	}

	renderTaskList(taskList)
}

func renderTaskList(taskList []entity.TaskList) {
	var rows [][]string

	re := lipgloss.NewRenderer(os.Stdout)
	var (
		purple    = lipgloss.Color("99")
		gray      = lipgloss.Color("#87de62")
		lightGray = lipgloss.Color("#de9e62")
		// HeaderStyle is the lipgloss style used for the table headers.
		HeaderStyle = re.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		// CellStyle is the base lipgloss style used for the table rows.
		CellStyle = re.NewStyle().Padding(0, 1).Width(5)
		// OddRowStyle is the lipgloss style used for odd-numbered table rows.
		OddRowStyle = CellStyle.Copy().Foreground(gray)
		// EvenRowStyle is the lipgloss style used for even-numbered table rows.
		EvenRowStyle = CellStyle.Copy().Foreground(lightGray)
		// BorderStyle is the lipgloss style used for the table border.
		BorderStyle = lipgloss.NewStyle().Foreground(purple)
	)

	columns := []string{
		"Name",
		"Role",
		"Priority",
		"Duration",
		"Done",
		"Left",
	}

	for _, task := range taskList {
		rows = append(rows, []string{
			task.Name,
			task.Role,
			strconv.Itoa(task.Priority),
			strconv.Itoa(task.TimeDuration),
			strconv.Itoa(task.TimeDone),
			strconv.Itoa(task.TimeDuration - task.TimeDone),
		})
	}

	t := table.New().
		Border(lipgloss.ThickBorder()).
		BorderStyle(BorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style

			switch {
			case row == 0:
				return HeaderStyle
			case row%2 == 0:
				style = EvenRowStyle
			default:
				style = OddRowStyle
			}
			// Make the second column a little wider.
			if col == 0 {
				style = style.Copy().Width(20)
			}
			if col == 1 {
				style = style.Copy().Width(10)
			}
			return style
		}).Headers(columns...).Rows(rows...)
	fmt.Println(t)
}
