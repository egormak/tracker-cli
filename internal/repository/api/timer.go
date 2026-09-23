package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"tracker_cli/internal/domain/entity"
)

type timerSetRequest struct {
	Count        int `json:"count"`
	TimeDuration int `json:"time_duration"`
}

type timerGlobalSetRequest struct {
	TimeScheduler int `json:"time_scheduler"`
}

type timerGlobalResponse struct {
	TimerGlobal int `json:"timer_global"`
}

// TimeDurationGet retrieves the current timer duration in minutes.
func TimeDurationGet() int {
	responseBody, err := sendRequest("GET", "/api/v1/timer/get", nil)
	if err != nil {
		slog.Error("request error", "error", err)
		return 0
	}
	defer responseBody.Close()

	var result entity.Timers
	if err := json.NewDecoder(responseBody).Decode(&result); err != nil {
		slog.Error("failed to decode response", "error", err)
		return 0
	}

	return result.Duration()
}

// TimeListSet sets the countdown timer duration on the backend.
func TimeListSet(count int) error {
	reqBody, err := json.Marshal(timerSetRequest{
		Count:        count,
		TimeDuration: count,
	})
	if err != nil {
		return fmt.Errorf("marshal timer set request: %w", err)
	}

	resp, err := sendRequest("POST", "/api/v1/timer/set", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("set timer list: %w", err)
	}
	defer resp.Close()

	return nil
}

// TimeDurationDel decreases/resets the countdown timer duration on the backend.
func TimeDurationDel(count int) error {
	reqBody, err := json.Marshal(timerSetRequest{
		Count:        count,
		TimeDuration: count,
	})
	if err != nil {
		return fmt.Errorf("marshal timer del request: %w", err)
	}

	resp, err := sendRequest("POST", "/api/v1/timer/del", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("delete timer: %w", err)
	}
	defer resp.Close()

	return nil
}

// SetGlobalTime sets the global scheduler timer on the backend.
func SetGlobalTime(timeScheduler int) error {
	reqBody, err := json.Marshal(timerGlobalSetRequest{
		TimeScheduler: timeScheduler,
	})
	if err != nil {
		return fmt.Errorf("marshal global timer request: %w", err)
	}

	resp, err := sendRequest("POST", "/api/v1/manage/timer/global", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("set global timer: %w", err)
	}
	defer resp.Close()

	return nil
}

// GetGlobalTime retrieves the global scheduler timer from the backend.
func GetGlobalTime() (int, error) {
	resp, err := sendRequest("GET", "/api/v1/manage/timer/global", nil)
	if err != nil {
		return 0, fmt.Errorf("get global timer: %w", err)
	}
	defer resp.Close()

	var result timerGlobalResponse
	if err := json.NewDecoder(resp).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode global timer response: %w", err)
	}

	return result.TimerGlobal, nil
}
