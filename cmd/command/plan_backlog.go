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

		batch, err := cmd.Flags().GetDuration("batch")
		if err != nil {
			return err
		}

		if batch > 0 {
			return plan.RunBacklogBatch(delay, batch)
		}

		return plan.RunBacklog(delay)
	},
}

func init() {
	planBacklogCmd.Flags().Duration("delay", 15*time.Second, "Delay before starting the task timer")
	planBacklogCmd.Flags().DurationP("batch", "b", 0, "Batch duration (e.g. 30m, 45m, 1h); stops when it is used up. Without it, runs until no backlog is left")
	planCmd.AddCommand(planBacklogCmd)
}
