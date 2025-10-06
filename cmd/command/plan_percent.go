package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tracker_cli/internal/service/plan"
	"tracker_cli/internal/service/procent"
)

var planPercentCmd = &cobra.Command{
	Use:   "percent",
	Short: "Manage percent-based planning",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlanPercent(cmd)
	},
}

func init() {
	planPercentRunCmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"start"},
		Short:   "Start the next task from the percent plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlanPercent(cmd)
		},
	}
	planPercentCmd.PersistentFlags().Duration("delay", 15*time.Second, "Delay before starting the task timer")

	planPercentSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Update the percent distribution for a role",
		RunE: func(cmd *cobra.Command, args []string) error {
			role, err := cmd.Flags().GetString("role")
			if err != nil {
				return err
			}
			rawValues, err := cmd.Flags().GetStringSlice("values")
			if err != nil {
				return err
			}

			percents, err := parsePercentValues(rawValues)
			if err != nil {
				return err
			}

			return procent.SetRolePercents(role, percents)
		},
	}
	planPercentSetCmd.Flags().String("role", "", "Role name to update")
	planPercentSetCmd.Flags().StringSlice("values", nil, "Percent values (comma-separated or repeated flag)")
	planPercentSetCmd.MarkFlagRequired("role")
	planPercentSetCmd.MarkFlagRequired("values")

	planPercentCmd.AddCommand(planPercentRunCmd)
	planPercentCmd.AddCommand(planPercentSetCmd)
	planCmd.AddCommand(planPercentCmd)
}

func runPlanPercent(cmd *cobra.Command) error {
	delay, err := cmd.Flags().GetDuration("delay")
	if err != nil {
		return err
	}

	return plan.RunPercent(delay)
}

func parsePercentValues(raw []string) ([]int, error) {
	var percents []int
	for _, token := range raw {
		for _, fragment := range strings.Split(token, ",") {
			fragment = strings.TrimSpace(fragment)
			if fragment == "" {
				continue
			}

			value, err := strconv.Atoi(fragment)
			if err != nil {
				return nil, fmt.Errorf("invalid percent value %q", fragment)
			}
			percents = append(percents, value)
		}
	}

	if len(percents) == 0 {
		return nil, fmt.Errorf("no percent values provided")
	}

	return percents, nil
}
