package command

import (
	"log/slog"
	"tracker_cli/internal/service/task_params"

	"github.com/spf13/cobra"
)

var taskAddCmd = &cobra.Command{
	Use:   "taskadd",
	Short: "Add a new task with a role",
	Run: func(cmd *cobra.Command, args []string) {
		taskName, _ := cmd.Flags().GetString("name")
		taskRole, _ := cmd.Flags().GetString("role")

		if taskName == "" {
			slog.Error("Task name is required")
			return
		}
		if taskRole == "" {
			slog.Error("Task role is required")
			return
		}

		slog.Info("Adding task", "name", taskName, "role", taskRole)
		task_params.SetTaskParams(taskName, 0, 0) // Example logic, adjust as needed
	},
}

func init() {
	taskAddCmd.Flags().StringP("name", "n", "", "Name of the task")
	taskAddCmd.Flags().StringP("role", "r", "", "Role of the task")
	taskAddCmd.MarkFlagRequired("name")
	taskAddCmd.MarkFlagRequired("role")

	rootCmd.AddCommand(taskAddCmd)
}
