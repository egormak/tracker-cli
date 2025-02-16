package manager

import (
	"log/slog"
	"os"
	"tracker_cli/internal/repository/api"

	"github.com/spf13/cobra"
)

func CleanRun(cmd *cobra.Command, args []string) {
	slog.Info("Run Clean")

	CleanData()
}

func CleanData() {

	result, err := api.CleanData()
	if err != nil {
		slog.Error("clean-data error", "error", err)
		os.Exit(1)
	}

	slog.Info(string(result.Message))
	slog.Info("Finish Data Clean")
}
