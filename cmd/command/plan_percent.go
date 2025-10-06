package command

import (
	"time"

	"github.com/spf13/cobra"

	"tracker_cli/internal/service/plan"
)

func newPlanPercentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "percent",
		Short: "Start the next task from the percent plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			delay, err := cmd.Flags().GetDuration("delay")
			if err != nil {
				return err
			}

			return plan.RunPercent(delay)
		},
	}

	cmd.Flags().Duration("delay", 15*time.Second, "Delay before starting the task timer")

	return cmd
}

func init() {
	planCmd.AddCommand(newPlanPercentCmd())
}
