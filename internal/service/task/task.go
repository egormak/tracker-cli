package task

import (
	"errors"
	"fmt"
	"log/slog"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"

	"github.com/spf13/cobra"
)

// ErrTaskCompleted indicates that no time remains for the task
var ErrTaskCompleted = fmt.Errorf("no time remaining for this task")

// ErrTaskAborted indicates the user aborted the task timer early and requested to stop the current plan.
var ErrTaskAborted = errors.New("task aborted by user")

func TaskRun(cmd *cobra.Command, args []string) error {
	taskName, err := cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("read task name flag: %w", err)
	}

	taskTime, err := cmd.Flags().GetInt("time")
	if err != nil {
		return fmt.Errorf("read task time flag: %w", err)
	}

	taskPercent, err := cmd.Flags().GetInt("percent")
	if err != nil {
		return fmt.Errorf("read task percent flag: %w", err)
	}

	sourceDay, err := cmd.Flags().GetString("source-day")
	if err != nil {
		return fmt.Errorf("read source day flag: %w", err)
	}

	previousDays, err := cmd.Flags().GetBool("previous-days")
	if err != nil {
		return fmt.Errorf("read previous-days flag: %w", err)
	}

	// If previous-days flag is set, use schedule-aware task retrieval
	if previousDays {
		return runTaskWithSchedule(taskName, taskTime, taskPercent, sourceDay)
	}

	taskApp, err := CreateTaskTimer(taskName, taskTime, taskPercent)
	if err != nil {
		if errors.Is(err, ErrTaskCompleted) {
			cmd.Printf("task %s has no remaining time for the selected percent\n", taskName)
			return nil
		}
		return err
	}

	// Set source day if provided
	if sourceDay != "" {
		taskApp.SourceDay = sourceDay
	}

	if err := taskApp.Run(); err != nil {
		if errors.Is(err, ErrTaskAborted) {
			slog.Info("task aborted by user", "task", taskName)
			return nil
		}
		return err
	}

	return nil
}

// CreateTaskTimer initializes a new TaskTimer object with the provided parameters
func CreateTaskTimer(name string, requestedDuration, percent int) (*TaskTimer, error) {
	taskParams := api.GetTaskParams(name)
	taskDone := api.StatisticTaskGet(name)

	duration, err := calculateDuration(taskParams, requestedDuration, percent, taskDone)
	if err != nil {
		return nil, fmt.Errorf("calculate duration: %w", err)
	}

	// Return a new TaskTimer object
	return &TaskTimer{
		Name:         name,
		Role:         api.TaskRoleGet(name),
		TimeDuration: duration,
		Percent:      percent,
	}, nil
}

// calculateDuration determines the appropriate time duration for the task
func calculateDuration(params entity.TaskParams, requested, percent, done int) (int, error) {
	if requested == 0 {
		return calculateDefaultDuration(params, percent, done)
	}
	return calculateRequestedDuration(params, requested, percent, done)

}

// calculateDefaultDuration handles the case when no specific duration is requested
func calculateDefaultDuration(params entity.TaskParams, percent, done int) (int, error) {
	apiDuration := api.TimeDurationGet()
	slog.Info("using default duration from API", "duration", apiDuration)

	if params == (entity.TaskParams{}) {
		return apiDuration, nil
	}

	timeLeft := calculateTimeLeft(params.Time, percent, done)
	if timeLeft <= 0 {
		return 0, ErrTaskCompleted
	}

	if timeLeft >= apiDuration {
		return apiDuration, nil
	}
	return timeLeft, nil
}

// calculateRequestedDuration handles the case when a specific duration is requested
func calculateRequestedDuration(params entity.TaskParams, requested, percent, done int) (int, error) {
	if params == (entity.TaskParams{}) {
		return requested, nil
	}
	fmt.Println("Time Duration: ", params.Time)
	timeLeft := calculateTimeLeft(params.Time, percent, done)
	if timeLeft <= 0 {
		return 0, ErrTaskCompleted
	}

	if timeLeft < requested {
		return timeLeft, nil
	}
	return requested, nil
}

// calculateTimeLeft calculates remaining time based on plan duration, percentage and time already spent
func calculateTimeLeft(planDuration, percent, done int) int {
	return (planDuration*percent)/100 - int(done)
}

// runTaskWithSchedule runs a task using schedule-aware lookup to find tasks from previous days
func runTaskWithSchedule(taskName string, requestedTime, requestedPercent int, explicitSourceDay string) error {
	// Get task info with schedule awareness (searches Monday to today)
	percent, timeLeft, sourceDay, err := api.GetTaskByNameSchedule(taskName)
	if err != nil {
		return fmt.Errorf("fetch scheduled task: %w", err)
	}

	// Use schedule-provided values or fall back to explicit flags
	finalPercent := percent
	if requestedPercent != 100 {
		// If user explicitly set percent, use that instead
		finalPercent = requestedPercent
	}

	finalSourceDay := sourceDay
	if explicitSourceDay != "" {
		// If user explicitly set source day, use that instead
		finalSourceDay = explicitSourceDay
	}

	// Log what we found
	if finalSourceDay != "" {
		slog.Info("schedule-aware task found (rollover)",
			"task", taskName,
			"percent", finalPercent,
			"time_left", timeLeft,
			"source_day", finalSourceDay)
	} else {
		slog.Info("schedule-aware task found",
			"task", taskName,
			"percent", finalPercent,
			"time_left", timeLeft)
	}

	// Determine duration to use
	var duration int
	if requestedTime > 0 {
		// Use user-requested time, but cap at timeLeft if available
		if timeLeft > 0 && timeLeft < requestedTime {
			duration = timeLeft
		} else {
			duration = requestedTime
		}
	} else if timeLeft > 0 {
		// Use timeLeft from schedule, but cap at default timer duration
		defaultDuration := api.TimeDurationGet()
		if timeLeft < defaultDuration {
			duration = timeLeft
		} else {
			duration = defaultDuration
		}
	} else {
		// Fall back to default duration
		duration = api.TimeDurationGet()
	}

	// Create task timer
	taskApp, err := CreateTaskTimer(taskName, duration, finalPercent)
	if err != nil {
		if errors.Is(err, ErrTaskCompleted) {
			slog.Info("task already completed for selected percent", "task", taskName)
			return nil
		}
		return fmt.Errorf("create task timer: %w", err)
	}

	// Set source day from schedule
	taskApp.SourceDay = finalSourceDay

	slog.Info("starting scheduled task",
		"task", taskName,
		"duration", duration,
		"percent", finalPercent)

	if err := taskApp.Run(); err != nil {
		if errors.Is(err, ErrTaskAborted) {
			slog.Info("task aborted by user", "task", taskName)
			return nil
		}
		return fmt.Errorf("run task: %w", err)
	}

	return nil
}
