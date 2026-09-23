package entity

import "testing"

func TestTimers_Duration(t *testing.T) {
	tests := []struct {
		name     string
		timer    Timers
		expected int
	}{
		{
			name:     "returns TimeDuration when TimeDuration is positive",
			timer:    Timers{TimeDuration: 25, Count: 0},
			expected: 25,
		},
		{
			name:     "returns TimeDuration when both are positive",
			timer:    Timers{TimeDuration: 25, Count: 10},
			expected: 25,
		},
		{
			name:     "falls back to Count when TimeDuration is 0",
			timer:    Timers{TimeDuration: 0, Count: 30},
			expected: 30,
		},
		{
			name:     "returns 0 when both are 0",
			timer:    Timers{TimeDuration: 0, Count: 0},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.timer.Duration(); got != tc.expected {
				t.Errorf("Duration() = %d, expected %d", got, tc.expected)
			}
		})
	}
}
