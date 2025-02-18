package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"tracker_cli/internal/domain/entity"
)

func TaskRoleGet(taskName string) string {

	var result entity.RoleAnswer

	responceBody, err := sendRequest("GET", fmt.Sprintf("%s?task_name=%s", "/api/v1/role/get", taskName), nil)

	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result.Role

}

func GetRoleRecords() map[string]int {

	result := make(map[string]int)

	responceBody, err := sendRequest("GET", "/api/v1/roles/records", nil)

	if err != nil {
		slog.Error("request error", "error", err)
		os.Exit(1)
	}

	defer responceBody.Close()

	err = json.NewDecoder(responceBody).Decode(&result)
	if err != nil {
		slog.Error("failed to decode response: %w", "error", err)
		os.Exit(1)
	}

	return result

}
