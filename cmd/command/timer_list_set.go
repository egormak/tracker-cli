package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"tracker_cli/internal/service/timer"
)

var timerListSetCmd = &cobra.Command{
	Use:   "timer-list-set",
	Short: "Generate a new timer list with a specified count",
	Long:  "Create or refresh the timer list on the tracker service by providing the number of timer slots to seed.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, err := cmd.Flags().GetInt("count")
		if err != nil {
			return err
		}
		if count <= 0 {
			return fmt.Errorf("count must be greater than zero")
		}

		timer.TimeListSet(count)
		return nil
	},
}

func init() {
	timerListSetCmd.Flags().IntP("count", "c", 0, "Number of timers to create in the list")
	timerListSetCmd.MarkFlagRequired("count")
	rootCmd.AddCommand(timerListSetCmd)
}
