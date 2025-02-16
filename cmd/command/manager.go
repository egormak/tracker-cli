package command

import (
	"tracker_cli/internal/service/manager"

	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Run Clean",
	Run:   manager.CleanRun,
}

func init() {
	rootCmd.AddCommand(CleanCmd)
}
