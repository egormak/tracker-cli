package command

import (
	"github.com/spf13/cobra"

	"tracker_cli/internal/service/role"
)

var roleRecheckCmd = &cobra.Command{
	Use:   "role-recheck",
	Short: "Recalculate and refresh role statistics",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		role.RoleRecheck()
	},
}

func init() {
	rootCmd.AddCommand(roleRecheckCmd)
}
