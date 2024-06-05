package manage

import (
	"log/slog"
	"os"
	"tracker_cli/internal/service/api/tracker_api"
)

func CleanData() {

	slog.Info("Start Data Clean")

	response, err := tracker_api.NewClient("POST", "/api/v1/records/clean", nil).Request()
	if err != nil {
		slog.Error("clean-data error", "error", err)
		os.Exit(1)
	}

	slog.Info(string(response))
	slog.Info("Finish Data Clean")
}
