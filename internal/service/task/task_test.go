package task

import (
	"errors"
	"testing"
	"tracker_cli/internal/domain/entity"
)

func TestCalculateDuration(t *testing.T) {
	tests := []struct {
		name      string
		params    entity.TaskParams
		requested int
		percent   int
		done      int
		want      int
		wantErr   error
	}{
		{
			name:      "Universal 10% stage for work task with requested 100m - capped to remaining 21m",
			params:    entity.TaskParams{Name: "work", Time: 270},
			requested: 100,
			percent:   10,
			done:      6,
			want:      21, // (270*10)/100 - 6 = 27 - 6 = 21
			wantErr:   nil,
		},
		{
			name:      "Universal 10% stage already completed for work task",
			params:    entity.TaskParams{Name: "work", Time: 270},
			requested: 100,
			percent:   10,
			done:      27,
			want:      0,
			wantErr:   ErrTaskCompleted,
		},
		{
			name:      "Universal 50% stage for english task with requested 15m - capped to remaining 10m",
			params:    entity.TaskParams{Name: "english", Time: 20},
			requested: 15,
			percent:   50,
			done:      0,
			want:      10, // (20*50)/100 - 0 = 10
			wantErr:   nil,
		},
		{
			name:      "Universal 100% stage for english task with requested 15m - within remaining 20m",
			params:    entity.TaskParams{Name: "english", Time: 20},
			requested: 15,
			percent:   100,
			done:      0,
			want:      15,
			wantErr:   nil,
		},
		{
			name:      "Task with no schedule params - uses requested duration",
			params:    entity.TaskParams{Name: "custom", Time: 0},
			requested: 45,
			percent:   100,
			done:      0,
			want:      45,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateDuration(tt.params, tt.requested, tt.percent, tt.done)
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
