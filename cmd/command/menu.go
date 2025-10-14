package command

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"tracker_cli/internal/service/menu"
	"tracker_cli/internal/service/task"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Launch the interactive menu and start a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Opening interactive menu")
		selectedTask := menu.RunMenu()

		if selectedTask == "" {
			cmd.Println("No task selected; exiting menu.")
			return nil
		}

		timeMinutes, err := cmd.Flags().GetInt("time")
		if err != nil {
			return fmt.Errorf("read time flag: %w", err)
		}

		percent, err := cmd.Flags().GetInt("percent")
		if err != nil {
			return fmt.Errorf("read percent flag: %w", err)
		}
		if percent <= 0 {
			return fmt.Errorf("percent must be greater than zero")
		}

		taskTimer, err := task.CreateTaskTimer(selectedTask, timeMinutes, percent)
		if err != nil {
			if errors.Is(err, task.ErrTaskCompleted) {
				cmd.Printf("Task %s has no remaining time.\n", selectedTask)
				return nil
			}
			return err
		}

		slog.Info("Starting task from menu selection", "task", selectedTask, "minutes", taskTimer.TimeDuration, "percent", percent)
		if err := taskTimer.Run(); err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				slog.Info("Task aborted from menu", "task", selectedTask)
				return nil
			}
			return err
		}
		return nil
	},
}

func init() {
	menuCmd.Flags().IntP("time", "t", 0, "Override time duration in minutes")
	menuCmd.Flags().IntP("percent", "p", 100, "Percent of planned time to run when no time override is provided")
	rootCmd.AddCommand(menuCmd)
}
