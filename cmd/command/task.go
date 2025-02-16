package command

import (
	"tracker_cli/internal/service/task"

	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run Task",
	Run:   task.TaskRun,
}

func init() {
	taskCmd.Flags().StringP("name", "n", "", "Task Name")
	taskCmd.Flags().IntP("time", "t", 0, "Time Duration")
	taskCmd.Flags().IntP("percent", "p", 100, "Percent of task time")

	taskCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(taskCmd)
}
