package command

import (
	"tracker_cli/internal/service"

	"github.com/spf13/cobra"
)

var statisticCmd = &cobra.Command{
	Use:   "statistic",
	Short: "Show statistic",
	Run:   service.StatisticShow,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("Show statistic")
	// },
}

func init() {
	rootCmd.AddCommand(statisticCmd)
}
