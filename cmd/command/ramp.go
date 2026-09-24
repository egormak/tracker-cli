package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tracker_cli/internal/domain/entity"
	"tracker_cli/internal/repository/api"
)

var rampCmd = &cobra.Command{
	Use:   "ramp",
	Short: "Linear warm-up ramp ladder commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRampStatus(cmd)
	},
}

var rampStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current warm-up ramp status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRampStatus(cmd)
	},
}

func runRampStatus(cmd *cobra.Command) error {
	status, err := api.GetRampStatus()
	if err != nil {
		return fmt.Errorf("failed to get ramp status: %w", err)
	}

	capStr := fmt.Sprintf("%d мин", status.CapMinutes)
	stepStr := fmt.Sprintf("[%d мин] (Шаг %d из %d)", status.CurrentStep, status.CurrentStep, status.CapMinutes)
	if status.IsCapped {
		stepStr = fmt.Sprintf("[%d мин] (Потолок достигнут!)", status.CurrentStep)
	}

	activeItems := append(status.Config.EnabledRoles, status.Config.EnabledTasks...)
	activeStr := strings.Join(activeItems, ", ")
	if activeStr == "" {
		activeStr = "нет"
	}

	excludedStr := strings.Join(status.Config.ExcludedTasks, ", ")
	if excludedStr == "" {
		excludedStr = "нет"
	}

	cmd.Println("⚡️ Warm-Up Ramp Ladder (Линейный разгон)")
	cmd.Printf("Текущий шаг:   %s\n", stepStr)
	cmd.Printf("Фокус сегодня: %d мин\n", status.TodayFocusMinutes)
	cmd.Printf("Потолок:       %s\n", capStr)
	cmd.Printf("Активно:       %s\n", activeStr)
	cmd.Printf("Исключения:    %s\n", excludedStr)
	cmd.Printf("Отдых fallback:%d мин\n\n", status.Config.DefaultRestFallback)
	cmd.Println("[r] Сбросить на 1м: tracker ramp reset")

	return nil
}

var rampResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset warm-up ramp step to 1 minute",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := api.ResetRamp()
		if err != nil {
			return fmt.Errorf("failed to reset ramp: %w", err)
		}
		cmd.Println("✅ Warm-Up Ramp reset to 1 minute")
		return nil
	},
}

var rampSetCapCmd = &cobra.Command{
	Use:   "set-cap <minutes>",
	Short: "Set maximum warm-up ramp cap in minutes (minimum 5)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minutes, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid minutes value: %s", args[0])
		}
		if minutes < 5 {
			return fmt.Errorf("cap must be at least 5 minutes")
		}

		currentStatus, err := api.GetRampStatus()
		var cfg entity.RampConfig
		if err == nil {
			cfg = currentStatus.Config
		}
		cfg.CapMinutes = minutes

		_, err = api.UpdateRampConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to update ramp cap: %w", err)
		}

		cmd.Printf("✅ Warm-Up Ramp cap set to %d minutes\n", minutes)
		return nil
	},
}

func init() {
	rampCmd.AddCommand(rampStatusCmd)
	rampCmd.AddCommand(rampResetCmd)
	rampCmd.AddCommand(rampSetCapCmd)
	rootCmd.AddCommand(rampCmd)
}
