package role

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"tracker_cli/config"
)

// RoleRecheck runs a database recheck for roles.
func RoleRecheck() {
	slog.Info("Run Recheck Roles Times")

	// Set the timeout to 15 seconds.
	timeout := time.Duration(15 * time.Second)

	// Create a new HTTP client with the specified timeout.
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Create a new GET request to the `/api/v1/role/recheck` route.
	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/role/recheck"), nil)

	if err != nil {
		slog.Error("request error", "error", err)
	}

	// Make the request.
	response, err := httpClient.Do(request)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	defer response.Body.Close()

	// Check the response status code.
	if response.StatusCode != http.StatusOK {
		slog.Error("request error", "status code", response.StatusCode)
	}

	slog.Info("Role recheck completed successfully")
}

func TaskRoleGet(taskName string) string {

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s?task_name=%s", config.TrackerDomain, "/api/v1/role/get", taskName), nil)
	if err != nil {
		slog.Error("error in request", "error", err)
	}
	resp, err := client.Do(request)
	if err != nil {
		slog.Error("error in request", "error", err)
	}

	if resp.StatusCode != 200 {
		slog.Error("request error", "status code", resp.StatusCode)
	}
	defer resp.Body.Close()

	var taskRole map[string]string
	err = json.NewDecoder(resp.Body).Decode(&taskRole)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
	}

	return taskRole["role"]
}
