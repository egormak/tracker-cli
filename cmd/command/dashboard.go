package command

import (
	"tracker_cli/internal/service/dashboard"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"tui", "dash"},
	Short:   "Launch the interactive TimeFlow TUI dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.RunDashboard()
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
