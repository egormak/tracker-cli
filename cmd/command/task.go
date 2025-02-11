package command

import (
	"tracker_cli/internal/service/task"

	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run Task",
	Run:   task.TaskRun,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("Show statistic")
	// },
}

func init() {
	taskCmd.Flags().String("name", "", "Task Name")
	taskCmd.Flags().Int("time", 0, "Time Duration")
	taskCmd.Flags().Int("percent", 100, "Percent of task time")

	rootCmd.AddCommand(taskCmd)
}
