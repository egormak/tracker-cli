package service

import (
	"log/slog"
	"os"
	"tracker_cli/internal/repository/api"
)

func CleanData() {

	slog.Info("Start Data Clean")

	result, err := api.CleanData()
	if err != nil {
		slog.Error("clean-data error", "error", err)
		os.Exit(1)
	}

	slog.Info(string(result.Message))
	slog.Info("Finish Data Clean")
}
