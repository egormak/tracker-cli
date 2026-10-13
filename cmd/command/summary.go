package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/digest"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show summaries",
}

var summaryWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show the weekly digest: time by role, strict tasks vs plan, choices vs schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		previous, _ := cmd.Flags().GetBool("previous")
		markdown, _ := cmd.Flags().GetBool("md")

		week := digest.WeekCurrent
		if previous {
			week = digest.WeekPrevious
		}

		d, err := api.GetWeeklyDigest(week)
		if err != nil {
			return fmt.Errorf("failed to fetch weekly digest: %w", err)
		}

		fmt.Fprint(cmd.OutOrStdout(), digest.Render(d, markdown))
		return nil
	},
}

func init() {
	summaryWeekCmd.Flags().Bool("previous", false, "Show the previous week (use on Monday before `tracker clean`)")
	summaryWeekCmd.Flags().Bool("md", false, "Print as markdown")
	summaryCmd.AddCommand(summaryWeekCmd)
	rootCmd.AddCommand(summaryCmd)
}
