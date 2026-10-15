package command

import (
	"github.com/spf13/cobra"
	"tracker_cli/internal/repository/api"
	debtrender "tracker_cli/internal/service/debt"
)

var debtCmd = &cobra.Command{
	Use:   "debt",
	Short: "Show this week's debt: the previous days' unfinished schedule and the week strip",
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := api.GetDebt()
		if err != nil {
			return err
		}
		cmd.Print(debtrender.Render(d))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(debtCmd)
}
