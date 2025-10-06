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

	taskApp, err := CreateTaskTimer(taskName, taskTime, taskPercent)
	if err != nil {
		if errors.Is(err, ErrTaskCompleted) {
			cmd.Printf("task %s has no remaining time for the selected percent\n", taskName)
			return nil
		}
		return err
	}

	return taskApp.Run()
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
