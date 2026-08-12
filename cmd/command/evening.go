package command

import (
	"log/slog"

	"github.com/spf13/cobra"
	"tracker_cli/internal/service/evening"
	"tracker_cli/internal/service/task"
)

var eveningCmd = &cobra.Command{
	Use:   "evening",
	Short: "Launch the Evening Catch-Up mode session",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Opening evening focus mode session")

		category, _ := cmd.Flags().GetString("category")
		sprintTime, _ := cmd.Flags().GetInt("time")
		skipTask, _ := cmd.Flags().GetString("skip")

		selectedTask, duration, err := evening.RunEveningSession(category, sprintTime, skipTask)
		if err != nil {
			return err
		}

		if selectedTask == "" {
			cmd.Println("No task selected for evening focus; exiting.")
			return nil
		}

		cmd.Printf("🚀 Starting Evening Focus Sprint on '%s' (%d min)...\n", selectedTask, duration)

		taskTimer, err := task.CreateTaskTimer(selectedTask, duration, 100)
		if err != nil {
			return err
		}

		if err := taskTimer.Run(); err != nil {
			if err.Error() == "task aborted" {
				return nil
			}
			return err
		}

		return nil
	},
}

func init() {
	eveningCmd.Flags().StringP("category", "c", "", "Filter tasks by category (e.g., learn, rest)")
	eveningCmd.Flags().IntP("time", "t", 20, "Sprint duration in minutes (15, 20, 30)")
	eveningCmd.Flags().StringP("skip", "s", "", "Task name to skip for tonight")

	rootCmd.AddCommand(eveningCmd)
}
