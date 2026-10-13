package manager

import (
	"fmt"
	"log/slog"
	"os"
	"time"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/digest"

	"github.com/spf13/cobra"
)

func CleanRun(cmd *cobra.Command, args []string) {
	slog.Info("Run Clean")

	// Last look at the week being wiped. A failed digest must not block the clean.
	if d, err := api.GetWeeklyDigest(digest.WeekBeforeClean(time.Now())); err != nil {
		slog.Warn("weekly digest unavailable, cleaning anyway", "error", err)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), digest.Render(d, false))
	}

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
