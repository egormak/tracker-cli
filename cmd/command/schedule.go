package command

import (
	"fmt"
	"strconv"
	"strings"
	"tracker_cli/internal/repository/api"
	"tracker_cli/internal/ui/theme"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:     "schedule",
	Aliases: []string{"sched"},
	Short:   "Manage weekly schedule and task allocations (Decision 0004)",
}

var scheduleAdjustCmd = &cobra.Command{
	Use:   "adjust [task_name] [delta_minutes]",
	Short: "Atomically adjust task duration in schedule (+/- minutes)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName := args[0]
		deltaStr := strings.TrimPrefix(args[1], "+")
		delta, err := strconv.Atoi(deltaStr)
		if err != nil {
			return fmt.Errorf("invalid delta minutes: %w", err)
		}

		day, _ := cmd.Flags().GetString("day")

		if err := api.UpdateScheduleTaskTime(taskName, delta, "adjust", day); err != nil {
			return fmt.Errorf("failed to adjust schedule task time: %w", err)
		}

		dayMsg := "today"
		if day != "" {
			dayMsg = day
		}
		cmd.Printf("✅ Adjusted task '%s' by %+d min on %s in weekly schedule\n", taskName, delta, dayMsg)
		return nil
	},
}

var scheduleSetCmd = &cobra.Command{
	Use:   "set [task_name] [target_minutes]",
	Short: "Atomically set task target duration in schedule",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName := args[0]
		target, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid target minutes: %w", err)
		}

		day, _ := cmd.Flags().GetString("day")

		if err := api.UpdateScheduleTaskTime(taskName, target, "set", day); err != nil {
			return fmt.Errorf("failed to set schedule task time: %w", err)
		}

		dayMsg := "today"
		if day != "" {
			dayMsg = day
		}
		cmd.Printf("✅ Set task '%s' duration to %d min on %s in weekly schedule\n", taskName, target, dayMsg)
		return nil
	},
}

var scheduleRolloverCmd = &cobra.Command{
	Use:     "rollover",
	Aliases: []string{"backlog"},
	Short:   "View rollover tasks carried over from previous weekdays",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := api.GetRolloverTasks()
		if err != nil {
			return fmt.Errorf("failed to fetch rollover tasks: %w", err)
		}

		if len(tasks) == 0 {
			cmd.Println("🎉 No rollover backlog tasks found. All previous schedule targets completed!")
			return nil
		}

		cmd.Println(theme.TitleStyle.Render("📋 ROLLOVER BACKLOG TASKS"))
		for _, t := range tasks {
			badge := theme.RoleBadge(t.Role)
			cmd.Printf("  %s %-18s Source: %-10s Remaining: %d min (%d%%)\n",
				badge, t.TaskName, t.SourceDay, t.RemainingTime, t.Percent)
		}
		return nil
	},
}

func init() {
	scheduleAdjustCmd.Flags().StringP("day", "d", "today", "Day to adjust: today, all, monday, tuesday, etc.")
	scheduleSetCmd.Flags().StringP("day", "d", "today", "Day to set: today, all, monday, tuesday, etc.")

	scheduleCmd.AddCommand(scheduleAdjustCmd)
	scheduleCmd.AddCommand(scheduleSetCmd)
	scheduleCmd.AddCommand(scheduleRolloverCmd)

	rootCmd.AddCommand(scheduleCmd)
}
