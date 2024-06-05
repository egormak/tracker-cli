package statistic

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"
	"tracker_cli/config"
	"tracker_cli/internal/service/task_params"
)

func StatisticShow() {

	slog.Info("Start Statistic Query")

	// Set Value
	resultRecords := make(map[string]map[string]int)
	resultRoleRecords := make(map[string]int)

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	// Get TaskRecords
	requestTaskRecords, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/records"), nil)
	respTaskRecords, err := client.Do(requestTaskRecords)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	if respTaskRecords.StatusCode != 200 {
		slog.Error("request error", "status code", respTaskRecords.StatusCode)
	}
	defer respTaskRecords.Body.Close()

	// Read the response body
	body, err := io.ReadAll(respTaskRecords.Body)
	if err != nil {
		slog.Error("can't read body", "error", err)
	}

	// Unmarshal JSON data into a struct
	err = json.Unmarshal(body, &resultRecords)
	if err != nil {
		slog.Error("can't unmarshal JSON", "error", err)
	}

	// Get RoleRecords
	requestRoleRecords, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/roles/records"), nil)
	respRoleRecords, err := client.Do(requestRoleRecords)
	if err != nil {
		slog.Error("request error", "error", err)
	}
	if respRoleRecords.StatusCode != 200 {
		slog.Error("request error", "status code", respTaskRecords.StatusCode)
	}
	defer respRoleRecords.Body.Close()

	// Read the response body
	body, err = io.ReadAll(respRoleRecords.Body)
	if err != nil {
		slog.Error("can't read body", "error", err)
	}

	// Unmarshal JSON data into a struct
	err = json.Unmarshal(body, &resultRoleRecords)
	if err != nil {
		slog.Error("can't unmarshal JSON", "error", err)
	}

	// TODO USE bubbletea
	// Show Information
	fmt.Println("####")
	fmt.Println("All Days was Done: ")
	for k, v := range resultRecords["all"] {
		fmt.Printf("Tasks: %s, Times: %d\n", k, v)
	}
	fmt.Println("\n####")
	fmt.Println("Yesterday was Done: ")
	for k, v := range resultRecords["yesterday"] {
		fmt.Printf("Tasks: %s, Times: %d\n", k, v)
	}
	fmt.Println("\n####")
	fmt.Println("For Today was Done: ")
	for k, v := range resultRecords["today"] {
		fmt.Printf("Tasks: %s, Times: %d\n", k, v)
	}
	fmt.Println("\n####")
	fmt.Println("Roles Info: ")
	for k, v := range resultRoleRecords {
		fmt.Printf("Roles: %s, Times: %d\n", k, v)
	}
}

func StatisticTaskShow(taskName string) {

	slog.Info("Begin Show Task")

	taskResult := StatisticTaskGet(taskName)

	timeSchedule := task_params.GetTaskParams(taskName).Time

	if timeSchedule == 0 {
		// Print informtaion
		slog.Info("Show Result", "task", taskName, "time_duration", taskResult)
	} else {
		// Print informtaion
		slog.Info("####")
		slog.Info("Show Result", "task", taskName, "time_duration", taskResult, "left", timeSchedule-taskResult)
		slog.Info("####")
	}
}

func StatisticTaskGet(taskName string) int {

	var taskResult struct {
		Status   string `json:"status"`
		TaskTime int    `json:"task_duration"`
	}

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s?task_name=%s", config.TrackerDomain, "/api/v1/record/task-day", taskName), nil)
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

	err = json.NewDecoder(resp.Body).Decode(&taskResult)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
	}

	return taskResult.TaskTime

}

func StatisticFullShow() {

	scheduledTimeToday := getScheduledTimeToday()
	if scheduledTimeToday == 0 {
		slog.Info("Time Limit for Today is not Set")
		return
	}

	completionTimeDone := statCompletionTimeDone()

	completionPercentage := statCompletionPercentage(completionTimeDone, scheduledTimeToday)

	timeLeft := scheduledTimeToday - completionTimeDone
	timePrediction := time.Now().Add(time.Minute * time.Duration(timeLeft))

	slog.Info("####")
	slog.Info("Percent Done", "percent", completionPercentage)
	slog.Info("Time Done", "time", completionTimeDone)
	slog.Info("Time prediction", "time", timePrediction.Format("15:04:05"))
}

func getScheduledTimeToday() int {

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/manage/timer/global"), nil)
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

	var timerGlobal map[string]int
	err = json.NewDecoder(resp.Body).Decode(&timerGlobal)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
	}

	return timerGlobal["timer_global"]
}

func statCompletionTimeDone() int {

	// Get
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", config.TrackerDomain, "/api/v1/roles/records/today"), nil)
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

	var timerDone map[string]int
	err = json.NewDecoder(resp.Body).Decode(&timerDone)
	if err != nil {
		slog.Error("failed to decode response: %w", err)
	}

	return timerDone["time_done"]

}

func statCompletionPercentage(completionTimeDone, scheduledTimeToday int) float64 {

	percent := float64(completionTimeDone) / float64(scheduledTimeToday) * 100
	round := math.Round(percent*100) / 100
	return round

}
