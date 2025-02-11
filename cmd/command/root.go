package command

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tracker",
	Short: "Tracker is a CLI tool for tracking time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to [your-cli-app-name]! Use --help for usage.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("error in command", "error", err)
		os.Exit(1)
	}
}
