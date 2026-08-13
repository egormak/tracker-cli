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

	percentSpecified := cmd.Flags().Changed("percent")

	// If previous-days flag is set, use schedule-aware task retrieval
	if previousDays {
		return runTaskWithSchedule(taskName, taskTime, taskPercent, sourceDay, percentSpecified)
	}

	taskApp, err := CreateTaskTimerWithPercentFlag(taskName, taskTime, taskPercent, percentSpecified)
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

// CreateTaskTimer initializes a new TaskTimer object with the provided parameters.
// Callers here pass a real, authoritative percent (not a CLI flag that may have been
// omitted), so the schedule cap and completion check are always enforced.
func CreateTaskTimer(name string, requestedDuration, percent int) (*TaskTimer, error) {
	return CreateTaskTimerWithPercentFlag(name, requestedDuration, percent, true)
}

// CreateTaskTimerWithPercentFlag initializes a TaskTimer object indicating whether the percent flag was explicitly specified
func CreateTaskTimerWithPercentFlag(name string, requestedDuration, percent int, percentSpecified bool) (*TaskTimer, error) {
	taskParams := api.GetTaskParams(name)
	taskDone := api.StatisticTaskGet(name)
	apiDuration := api.TimeDurationGet()

	duration, err := calculateDuration(taskParams, requestedDuration, percent, taskDone, apiDuration, percentSpecified)
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

// calculateDuration determines the appropriate time duration for any task based on schedule, percent, time done, and explicit flags
func calculateDuration(params entity.TaskParams, requested, percent, done, apiDuration int, percentSpecified bool) (int, error) {
	if params == (entity.TaskParams{}) || params.Time == 0 {
		if requested > 0 {
			return requested, nil
		}
		return apiDuration, nil
	}

	timeLeft := calculateTimeLeft(params.Time, percent, done)

	if percentSpecified {
		if timeLeft <= 0 {
			return 0, ErrTaskCompleted
		}
		if requested > 0 {
			if timeLeft < requested {
				return timeLeft, nil
			}
			return requested, nil
		}
		if apiDuration <= timeLeft {
			return apiDuration, nil
		}
		return timeLeft, nil
	}

	// When percent was NOT explicitly specified (e.g. tracker task -n video -t 25):
	if requested > 0 {
		return requested, nil
	}

	if timeLeft <= 0 {
		return 0, ErrTaskCompleted
	}

	if apiDuration <= timeLeft {
		return apiDuration, nil
	}
	return timeLeft, nil
}

// calculateTimeLeft calculates remaining time based on plan duration, percentage and time already spent
func calculateTimeLeft(planDuration, percent, done int) int {
	return (planDuration*percent)/100 - int(done)
}

// runTaskWithSchedule runs a task using schedule-aware lookup to find tasks from previous days
func runTaskWithSchedule(taskName string, requestedTime, requestedPercent int, explicitSourceDay string, percentSpecified bool) error {
	// Get task info with schedule awareness (searches Monday to today)
	percent, timeLeft, sourceDay, err := api.GetTaskByNameSchedule(taskName)
	if err != nil {
		return fmt.Errorf("fetch scheduled task: %w", err)
	}

	// Use schedule-provided values or fall back to explicit flags
	finalPercent := percent
	if percentSpecified {
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
	taskApp, err := CreateTaskTimerWithPercentFlag(taskName, duration, finalPercent, percentSpecified)
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
