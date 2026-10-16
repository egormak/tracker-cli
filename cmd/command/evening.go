package command

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"
	debtrender "tracker_cli/internal/service/debt"
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
		comboTime, _ := cmd.Flags().GetInt("combo-time")

		if batch, _ := cmd.Flags().GetBool("batch"); batch {
			return runEveningBatch(cmd, category, skipTask)
		}

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
			if comboTime > 0 {
				comboDuration = comboTime
			} else if comboFlag && cmd.Flags().Changed("time") {
				comboDuration = sprintTime
			} else if duration > 0 {
				comboDuration = duration / 2
			}
			if comboDuration <= 0 {
				comboDuration = 10
			}

			return evening.RunCombo(cmd.OutOrStderr(), topCandidates, comboDuration)
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

		taskTimer.Origin = "evening"
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
	eveningCmd.Flags().BoolP("combo", "C", false, "Launch sequential combo chain across top-3 candidates")
	eveningCmd.Flags().IntP("combo-time", "d", 0, "Sprint duration for each step in combo chain (e.g. 10, 15, 20)")
	eveningCmd.Flags().BoolP("batch", "b", false, "Run an evening batch: the anchor and the next candidates by debt, one sprint each (-t, default 15)")
	eveningCmd.Flags().StringP("anchor", "a", "", "Batch: first task; without it the picker asks (Enter picks the anchor)")
	eveningCmd.Flags().Int("debt", 0, "Percent (0-100) of every sprint that pays the task's previous-days debt")
	eveningCmd.Flags().Int("tasks", 3, "Batch: number of tasks, from 1 to the anchor plus tonight's candidates")

	rootCmd.AddCommand(eveningCmd)
}

// runEveningBatch runs `tracker evening --batch`: the anchor and the next evening candidates
// (ranked by the server by debt to date) as one sprint each.
func runEveningBatch(cmd *cobra.Command, category, skipTask string) error {
	sprint := 15
	if cmd.Flags().Changed("time") {
		sprint, _ = cmd.Flags().GetInt("time")
	}
	anchor, _ := cmd.Flags().GetString("anchor")
	debtShare, _ := cmd.Flags().GetInt("debt")
	tasks, _ := cmd.Flags().GetInt("tasks")
	if sprint <= 0 {
		return fmt.Errorf("sprint must be positive, got %d", sprint)
	}
	if debtShare < 0 || debtShare > 100 {
		return fmt.Errorf("--debt must be between 0 and 100, got %d", debtShare)
	}
	if tasks < 1 {
		return fmt.Errorf("--tasks must be at least 1, got %d", tasks)
	}

	if anchor == "" {
		selected, _, isCombo, _, err := evening.RunEveningSession(category, sprint, skipTask)
		if err != nil {
			return err
		}
		if selected == "" && !isCombo {
			cmd.Println("No anchor selected; exiting.")
			return nil
		}
		anchor = selected // empty after "c": the first candidate leads
	} else if skipTask != "" {
		_, _ = api.SkipEveningTask(skipTask, category, sprint)
	}

	focus, err := api.GetEveningFocus(category, sprint)
	if err != nil {
		return fmt.Errorf("failed to fetch evening focus: %w", err)
	}
	capacity := evening.BatchCapacity(focus.Candidates, anchor)
	if capacity == 0 {
		cmd.Println("All weekly tasks completed.")
		return nil
	}
	if tasks > capacity {
		// The default shrinks to what is available; an explicit --tasks must fit.
		if cmd.Flags().Changed("tasks") {
			return fmt.Errorf("--tasks must be between 1 and %d (tasks available tonight), got %d", capacity, tasks)
		}
		tasks = capacity
	}
	batch := evening.BuildBatch(focus.Candidates, anchor, tasks)

	names := make([]string, len(batch))
	for i, c := range batch {
		names[i] = c.TaskName
	}
	plan := fmt.Sprintf("🧩 Батч: %s · по %d мин", strings.Join(names, " → "), sprint)
	if debtShare > 0 {
		plan += fmt.Sprintf(" · %d%% в долг", debtShare)
	}
	cmd.Println(plan)

	if err := evening.RunBatch(cmd.OutOrStderr(), batch, sprint, debtShare); err != nil {
		return err
	}
	if d, err := api.GetDebt(); err == nil {
		cmd.Println(debtrender.Line(d))
		cmd.Println(debtrender.WeekStrip(d))
	}
	return nil
}
