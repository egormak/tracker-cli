package command

import (
	"github.com/spf13/cobra"

	"tracker_cli/internal/service/rest"
)

var restSpendCmd = &cobra.Command{
	Use:   "rest-spend [duration]",
	Short: "Set how much time you spent on rest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := cmd.Flags().GetInt("duration")
		if err != nil {
			return err
		}
		rest.RestSpend(duration)
		return nil
	},
}

func init() {
	restSpendCmd.Flags().IntP("duration", "d", 0, "Duration of rest in minutes")
	restSpendCmd.MarkFlagRequired("duration")
	rootCmd.AddCommand(restSpendCmd)
}
