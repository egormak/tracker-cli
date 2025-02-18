package api

import (
	"encoding/json"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
)

func TimeDurationGet() int {

	var result entity.Timers

	responceBody, err := sendRequest("GET", "/api/v1/timer/get", nil)

	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result.TimeDuration

}
