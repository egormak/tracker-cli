package command

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/service/evening"
	"tracker_cli/internal/service/task"
)

var eveningCmd = &cobra.Command{
	Use:   "evening",
	Short: "Launch the Evening Catch-Up mode session",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Opening evening focus mode session")

		category, _ := cmd.Flags().GetString("category")
		sprintTime, _ := cmd.Flags().GetInt("time")
		skipTask, _ := cmd.Flags().GetString("skip")
		comboFlag, _ := cmd.Flags().GetBool("combo")

		var selectedTask string
		var duration int
		var isCombo bool
		var topCandidates []entity.EveningFocusCandidate

		if comboFlag {
			if sprintTime <= 0 {
				sprintTime = 20
			}
			if skipTask != "" {
				_, _ = api.SkipEveningTask(skipTask, category, sprintTime)
			}
			focus, err := api.GetEveningFocus(category, sprintTime)
			if err != nil {
				return fmt.Errorf("failed to fetch evening focus: %w", err)
			}
			topCandidates = evening.GetTopCandidates(focus)
			duration = sprintTime
			isCombo = true
		} else {
			var err error
			selectedTask, duration, isCombo, topCandidates, err = evening.RunEveningSession(category, sprintTime, skipTask)
			if err != nil {
				return err
			}
		}

		if isCombo {
			if len(topCandidates) == 0 {
				cmd.Println("All weekly tasks completed.")
				return nil
			}

			comboDuration := 10
			if duration > 0 {
				comboDuration = duration / 2
			}
			if comboDuration <= 0 {
				comboDuration = 10
			}

			for i, taskCandidate := range topCandidates {
				cmd.Printf("⚡️ [Combo %d/%d] Starting sprint on '%s' (%d min)...\n", i+1, len(topCandidates), taskCandidate.TaskName, comboDuration)

				taskTimer, err := task.CreateTaskTimerWithPercentFlag(taskCandidate.TaskName, comboDuration, 100, false)
				if err != nil {
					if errors.Is(err, task.ErrTaskCompleted) {
						cmd.Printf("Task %s has no remaining time.\n", taskCandidate.TaskName)
						continue
					}
					return err
				}

				if err := taskTimer.Run(); err != nil {
					if errors.Is(err, task.ErrTaskAborted) {
						return nil
					}
					return err
				}

				cmd.Printf("✅ Sprint on '%s' completed!\n", taskCandidate.TaskName)
			}

			cmd.Println("🎉 Evening Combo Chain completed successfully!")
			return nil
		}

		if selectedTask == "" {
			cmd.Println("No task selected for evening focus; exiting.")
			return nil
		}

		cmd.Printf("🚀 Starting Evening Focus Sprint on '%s' (%d min)...\n", selectedTask, duration)

		// Evening sprint duration is derived from weekly deficit, not today's daily
		// target, so percentSpecified=false bypasses the daily-target timeLeft cap
		// (flexible/background tasks have tiny daily targets often already spent today).
		taskTimer, err := task.CreateTaskTimerWithPercentFlag(selectedTask, duration, 100, false)
		if err != nil {
			if errors.Is(err, task.ErrTaskCompleted) {
				cmd.Printf("Task %s has no remaining time.\n", selectedTask)
				return nil
			}
			return err
		}

		if err := taskTimer.Run(); err != nil {
			if errors.Is(err, task.ErrTaskAborted) {
				return nil
			}
			return err
		}

		return nil
	},
}

func init() {
	eveningCmd.Flags().StringP("category", "c", "", "Filter tasks by category (e.g., learn, rest)")
	eveningCmd.Flags().IntP("time", "t", 20, "Sprint duration in minutes (15, 20, 30)")
	eveningCmd.Flags().StringP("skip", "s", "", "Task name to skip for tonight")
	eveningCmd.Flags().BoolP("combo", "C", false, "Launch sequential combo chain 3x10m across top-3 candidates")

	rootCmd.AddCommand(eveningCmd)
}
