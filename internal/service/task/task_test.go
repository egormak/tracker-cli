package task

import (
	"errors"
	"testing"
	"tracker_cli/internal/domain/entity"
)

func TestCalculateDuration(t *testing.T) {
	tests := []struct {
		name             string
		params           entity.TaskParams
		requested        int
		percent          int
		done             int
		apiDuration      int
		percentSpecified bool
		want             int
		wantErr          error
	}{
		{
			name:             "Explicit requested duration -t 100 with -p 50 for work task (params.Time=270, done=125)",
			params:           entity.TaskParams{Name: "work", Time: 270},
			requested:        100,
			percent:          50,
			done:             125,
			percentSpecified: true,
			want:             10, // target = (270*50)/100 = 135; timeLeft = 135-125 = 10; min(100, 10) = 10
			wantErr:          nil,
		},
		{
			name:             "Explicit requested duration -t 100 when target percent already completed (params.Time=270, done=140)",
			params:           entity.TaskParams{Name: "work", Time: 270},
			requested:        100,
			percent:          50,
			done:             140,
			percentSpecified: true,
			want:             0,
			wantErr:          ErrTaskCompleted,
		},
		{
			name:             "Explicit requested duration -t 25 for soft schedule video task (no explicit -p flag)",
			params:           entity.TaskParams{Name: "video", Time: 5},
			requested:        25,
			percent:          100,
			done:             6,
			percentSpecified: false,
			want:             25, // manual duration without -p runs for requested duration
			wantErr:          nil,
		},
		{
			name:             "No requested duration (requested=0) - caps default session to remaining schedule time",
			params:           entity.TaskParams{Name: "english", Time: 20},
			requested:        0,
			percent:          100,
			done:             15,
			apiDuration:      5,
			percentSpecified: false,
			want:             5, // min(apiDuration=5, remaining=5) = 5
			wantErr:          nil,
		},
		{
			name:             "No requested duration (requested=0) - returns ErrTaskCompleted when schedule is complete",
			params:           entity.TaskParams{Name: "english", Time: 20},
			requested:        0,
			percent:          100,
			done:             20,
			percentSpecified: false,
			want:             0,
			wantErr:          ErrTaskCompleted,
		},
		{
			name:             "Task with no schedule params - uses requested duration",
			params:           entity.TaskParams{Name: "custom", Time: 0},
			requested:        45,
			percent:          100,
			done:             0,
			percentSpecified: false,
			want:             45,
			wantErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateDuration(tt.params, tt.requested, tt.percent, tt.done, tt.apiDuration, tt.percentSpecified)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("calculateDuration() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("calculateDuration() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("calculateDuration() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateTimeLeft(t *testing.T) {
	tests := []struct {
		planDuration int
		percent      int
		done         int
		want         int
	}{
		{planDuration: 270, percent: 10, done: 6, want: 21},
		{planDuration: 270, percent: 10, done: 27, want: 0},
		{planDuration: 270, percent: 100, done: 6, want: 264},
		{planDuration: 20, percent: 50, done: 0, want: 10},
		{planDuration: 20, percent: 100, done: 20, want: 0},
	}

	for _, tt := range tests {
		got := calculateTimeLeft(tt.planDuration, tt.percent, tt.done)
		if got != tt.want {
			t.Errorf("calculateTimeLeft(%d, %d, %d) = %d, want %d", tt.planDuration, tt.percent, tt.done, got, tt.want)
		}
	}
}
