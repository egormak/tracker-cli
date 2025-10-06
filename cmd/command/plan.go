package command

import "github.com/spf13/cobra"

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Work with planned tasks",
}

func init() {
	rootCmd.AddCommand(planCmd)
}
