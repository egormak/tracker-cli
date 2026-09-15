package command

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"tracker_cli/internal/service/task_params"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var taskAddCmd = &cobra.Command{
	Use:     "taskadd",
	Aliases: []string{"add-task", "new-task"},
	Short:   "Add a new task with a role and parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName, _ := cmd.Flags().GetString("name")
		taskRole, _ := cmd.Flags().GetString("role")
		taskTime, _ := cmd.Flags().GetInt("time")
		taskPriority, _ := cmd.Flags().GetInt("priority")

		// If required flags are omitted, launch interactive Huh form
		if taskName == "" || taskRole == "" {
			var timeStr string
			var priorityStr string

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Task Name").
						Description("Enter the name of the new task").
						Value(&taskName).
						Validate(func(str string) error {
							if strings.TrimSpace(str) == "" {
								return errors.New("task name cannot be empty")
							}
							return nil
						}),
					huh.NewSelect[string]().
						Title("Task Category / Role").
						Options(
							huh.NewOption("Work 🟠", "work"),
							huh.NewOption("Learn / Study 🔵", "learn"),
							huh.NewOption("Rest 🟢", "rest"),
						).
						Value(&taskRole),
					huh.NewInput().
						Title("Planned Daily Duration (min)").
						Value(&timeStr).
						Placeholder("25"),
					huh.NewInput().
						Title("Priority (1-10)").
						Value(&priorityStr).
						Placeholder("1"),
				),
			)

			if err := form.Run(); err != nil {
				return fmt.Errorf("task add form cancelled: %w", err)
			}

			if timeStr != "" {
				if t, err := strconv.Atoi(strings.TrimSpace(timeStr)); err == nil {
					taskTime = t
				}
			}
			if priorityStr != "" {
				if p, err := strconv.Atoi(strings.TrimSpace(priorityStr)); err == nil {
					taskPriority = p
				}
			}
		}

		slog.Info("Adding task", "name", taskName, "role", taskRole, "time", taskTime, "priority", taskPriority)
		task_params.SetTaskParams(taskName, taskTime, taskPriority)
		cmd.Printf("✅ Successfully configured task '%s' (role: %s, target: %d min, priority: %d)\n", taskName, taskRole, taskTime, taskPriority)
		return nil
	},
}

func init() {
	taskAddCmd.Flags().StringP("name", "n", "", "Name of the task")
	taskAddCmd.Flags().StringP("role", "r", "", "Role of the task (work, learn, rest)")
	taskAddCmd.Flags().IntP("time", "t", 0, "Planned duration in minutes")
	taskAddCmd.Flags().IntP("priority", "P", 1, "Priority level (1-10)")

	rootCmd.AddCommand(taskAddCmd)
}

