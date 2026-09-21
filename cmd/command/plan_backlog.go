package command

import (
	"time"

	"github.com/spf13/cobra"
	"tracker_cli/internal/service/plan"
)

var planBacklogCmd = &cobra.Command{
	Use:     "backlog",
	Aliases: []string{"catchup", "game"},
	Short:   "Start sequencing through all deficit tasks (past and today)",
	RunE: func(cmd *cobra.Command, args []string) error {
		delay, err := cmd.Flags().GetDuration("delay")
		if err != nil {
			return err
		}

		restLimit, err := cmd.Flags().GetInt("rest-limit")
		if err != nil {
			return err
		}

		batch, err := cmd.Flags().GetDuration("batch")
		if err != nil {
			return err
		}

		if batch > 0 {
			return plan.RunBacklogBatch(delay, restLimit, batch)
		}

		return plan.RunBacklog(delay, restLimit)
	},
}

func init() {
	planBacklogCmd.Flags().Duration("delay", 15*time.Second, "Delay before starting the task timer")
	planBacklogCmd.Flags().IntP("rest-limit", "r", -1, "Maximum rest minutes before stopping; negative disables continuous mode")
	planBacklogCmd.Flags().DurationP("batch", "b", 0, "Batch duration (e.g. 30m, 45m, 1h); runs tasks until batch time is reached")
	planCmd.AddCommand(planBacklogCmd)
}
