package command

import (
	"tracker_cli/internal/service/timer"

	"github.com/spf13/cobra"
)

var timerRecheckCmd = &cobra.Command{
	Use:   "timer-recheck",
	Short: "Recheck timers and refresh timer list",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		timer.TimerRecheck()
	},
}

func init() {
	rootCmd.AddCommand(timerRecheckCmd)
}
