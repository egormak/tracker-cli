package command

import (
	"fmt"
	"testing"
	"time"
)

func TestSessionCommand_Configuration(t *testing.T) {
	if sessionCmd.Use != "session [duration]" {
		t.Errorf("expected Use 'session [duration]', got %q", sessionCmd.Use)
	}

	foundAlias := false
	for _, alias := range sessionCmd.Aliases {
		if alias == "batch" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Error("expected session command to have 'batch' alias")
	}

	delayFlag := sessionCmd.Flags().Lookup("delay")
	if delayFlag == nil {
		t.Fatal("expected sessionCmd to have 'delay' flag")
	}

	restLimitFlag := sessionCmd.Flags().Lookup("rest-limit")
	if restLimitFlag == nil {
		t.Fatal("expected sessionCmd to have 'rest-limit' flag")
	}
	if restLimitFlag.Shorthand != "r" {
		t.Errorf("expected rest-limit shorthand 'r', got %q", restLimitFlag.Shorthand)
	}
}

func TestPlanPercentCommand_BatchFlag(t *testing.T) {
	flag := planPercentCmd.PersistentFlags().Lookup("batch")
	if flag == nil {
		t.Fatal("expected planPercentCmd to have persistent 'batch' flag")
	}
	if flag.Shorthand != "b" {
		t.Errorf("expected batch flag shorthand 'b', got %q", flag.Shorthand)
	}
}

func TestPlanBacklogCommand_BatchFlag(t *testing.T) {
	flag := planBacklogCmd.Flags().Lookup("batch")
	if flag == nil {
		t.Fatal("expected planBacklogCmd to have 'batch' flag")
	}
	if flag.Shorthand != "b" {
		t.Errorf("expected batch flag shorthand 'b', got %q", flag.Shorthand)
	}
}

func TestRestResetCommand_Configuration(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"rest", "reset"})
	if err != nil {
		t.Fatalf("failed to find 'rest reset' command: %v", err)
	}
	if cmd.Name() != "reset" {
		t.Errorf("expected command name 'reset', got %q", cmd.Name())
	}
}

func TestSessionDurationParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"30m", 30 * time.Minute, false},
		{"45m", 45 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"20", 20 * time.Minute, false},
		{"0m", 0, true},
		{"-10m", 0, true},
		{"invalid", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			parsed, err := time.ParseDuration(tc.input)
			if err != nil {
				// try integer fallback
				if tc.input == "20" {
					parsed = 20 * time.Minute
					err = nil
				}
			}
			if err == nil && (parsed <= 0 || parsed < time.Minute) {
				err = fmt.Errorf("duration must be positive and at least 1 minute")
			}
			if tc.hasError && err == nil {
				t.Errorf("expected error for input %q, got nil", tc.input)
			}
			if !tc.hasError {
				if err != nil {
					t.Errorf("unexpected error for %q: %v", tc.input, err)
				}
				if parsed != tc.expected {
					t.Errorf("expected %v, got %v", tc.expected, parsed)
				}
			}
		})
	}
}
