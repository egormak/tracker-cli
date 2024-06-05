package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"tracker_cli/config"
)

func TelegramStartSend(taskName string) int {

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	var body struct {
		Name string `json:"task_name"`
	}

	body.Name = taskName

	json_data, err := json.Marshal(&body)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
		os.Exit(1)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/manage/telegram/start"), bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}

	var result struct {
		MsgID int `json:"msg_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
		os.Exit(1)
	}
	return result.MsgID
}

func TelegramStopSend(taskName string, msgID int, timeDone int, timeEnd string) {

	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	var body struct {
		Name     string `json:"task_name"`
		MsgID    int    `json:"msg_id"`
		TimeDone int    `json:"time_done"`
		TimeEnd  string `json:"time_end"`
	}

	body.Name = taskName
	body.MsgID = msgID
	body.TimeDone = timeDone
	body.TimeEnd = timeEnd

	json_data, err := json.Marshal(&body)
	if err != nil {
		slog.Error("can't marshal JSON", "error", err)
		os.Exit(1)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/manage/telegram/stop"), bytes.NewBuffer(json_data))
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
		os.Exit(1)
	}

}
