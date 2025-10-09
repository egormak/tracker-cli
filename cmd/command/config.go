package command

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"tracker_cli/internal/service/task_params"
	"tracker_cli/internal/service/timer"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure global timer or per-task parameters",
	Long:  "Update the global scheduler time or set task-specific timer and priority parameters.",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName, err := cmd.Flags().GetString("task")
		if err != nil {
			return err
		}

		timeMinutes, err := cmd.Flags().GetInt("time")
		if err != nil {
			return err
		}
		if timeMinutes <= 0 {
			return fmt.Errorf("time must be greater than zero minutes")
		}

		if taskName == "" {
			slog.Info("Setting global timer", "minutes", timeMinutes)
			timer.SetGlobalTime(timeMinutes)
			return nil
		}

		priority, err := cmd.Flags().GetInt("priority")
		if err != nil {
			return err
		}
		if priority <= 0 {
			return fmt.Errorf("priority must be greater than zero when setting task parameters")
		}

		slog.Info("Setting task parameters", "task", taskName, "minutes", timeMinutes, "priority", priority)
		task_params.SetTaskParams(taskName, timeMinutes, priority)
		return nil
	},
}

func init() {
	configCmd.Flags().StringP("task", "n", "", "Task name to configure")
	configCmd.Flags().IntP("time", "t", 0, "Time duration in minutes")
	configCmd.Flags().IntP("priority", "p", 0, "Priority value for the task")
	rootCmd.AddCommand(configCmd)
}
