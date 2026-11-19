package entity

import "testing"

func TestDashboardStateTask(t *testing.T) {
	state := DashboardState{
		TodayTasks: []TaskList{
			{Name: "work", Role: "work", TimeDuration: 270, TimeDone: 90, Priority: 1},
			{Name: "english", Role: "learn", TimeDuration: 30, TimeDone: 0, Priority: 2},
		},
	}

	tests := []struct {
		name     string
		query    string
		wantOK   bool
		wantDone int
	}{
		{"found", "work", true, 90},
		{"missing", "missing", false, 0},
		{"case sensitive", "Work", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := state.Task(tt.query)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got.TimeDone != tt.wantDone {
				t.Errorf("TimeDone = %d, want %d", got.TimeDone, tt.wantDone)
			}
		})
	}
}
