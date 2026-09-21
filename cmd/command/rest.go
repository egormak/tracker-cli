package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/rest"
)

var restSpendCmd = &cobra.Command{
	Use:   "rest-spend",
	Short: "Set how much time you spent on rest",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := cmd.Flags().GetInt("duration")
		if err != nil {
			return err
		}
		rest.RestSpend(duration)
		return nil
	},
}

var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Rest management commands",
}

var restResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset daily rest balance to 0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := api.ResetRest(); err != nil {
			return fmt.Errorf("failed to reset rest: %w", err)
		}
		fmt.Println("Rest balance has been reset to 0.")
		return nil
	},
}

func init() {
	restSpendCmd.Flags().IntP("duration", "d", 0, "Duration of rest in minutes")
	restSpendCmd.MarkFlagRequired("duration")
	rootCmd.AddCommand(restSpendCmd)

	restCmd.AddCommand(restResetCmd)
	rootCmd.AddCommand(restCmd)
}
