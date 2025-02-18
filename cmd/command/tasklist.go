package command

import (
	"tracker_cli/internal/service/statistic"

	"github.com/spf13/cobra"
)

var tasklistCmd = &cobra.Command{
	Use:   "tasklist",
	Short: "Show the list of tasks",
	Run: func(cmd *cobra.Command, args []string) {
		statistic.ShowTaskNameList()
	},
}

func init() {
	rootCmd.AddCommand(tasklistCmd)
}
