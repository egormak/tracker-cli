package command

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"tracker_cli/internal/service/plan"
)

var sessionCmd = &cobra.Command{
	Use:     "session [duration]",
	Aliases: []string{"batch"},
	Short:   "Run a focused batch session using the schedule-aware percent plan",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		duration := 30 * time.Minute
		if len(args) > 0 {
			parsed, err := time.ParseDuration(args[0])
			if err != nil {
				if mins, err2 := strconv.Atoi(args[0]); err2 == nil && mins > 0 {
					parsed = time.Duration(mins) * time.Minute
				} else {
					return fmt.Errorf("invalid session duration %q: %w", args[0], err)
				}
			}
			duration = parsed
		}
		if duration <= 0 {
			return fmt.Errorf("session duration must be positive")
		}
		if duration < time.Minute {
			return fmt.Errorf("batch duration must be at least 1 minute")
		}

		delay, err := cmd.Flags().GetDuration("delay")
		if err != nil {
			return err
		}

		restLimit, err := cmd.Flags().GetInt("rest-limit")
		if err != nil {
			return err
		}

		return plan.RunPercentBatch(delay, restLimit, duration, true)
	},
}

func init() {
	sessionCmd.Flags().Duration("delay", 15*time.Second, "Delay before starting each task timer")
	sessionCmd.Flags().IntP("rest-limit", "r", -1, "Maximum rest minutes before stopping")
	rootCmd.AddCommand(sessionCmd)
}
