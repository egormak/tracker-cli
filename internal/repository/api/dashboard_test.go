package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"tracker_cli/internal/domain/entity"
)

const dashboardBodyA = `{"running_task": null, "today_tasks": [{"name": "work", "role": "work", "time_duration": 270, "time_done": 90, "priority": 1}, {"name": "english", "role": "learn", "time_duration": 30, "time_done": 0, "priority": 2}], "evening_focus": null, "total_planned": 300, "total_done": 90, "completion_pct": 30}`

func fakeDashboard(status int, body string, paths *[]string) roundTripFunc {
	return func(req *http.Request) (*http.Response, error) {
		if paths != nil {
			*paths = append(*paths, req.Method+" "+req.URL.Path)
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}
}

func TestGetDashboardState(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var paths []string
		client.Transport = fakeDashboard(http.StatusOK, dashboardBodyA, &paths)

		state, err := GetDashboardState()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 1 || paths[0] != "GET /api/v1/dashboard/state" {
			t.Errorf("unexpected requests: %v", paths)
		}
		if len(state.TodayTasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(state.TodayTasks))
		}
		want := entity.TaskList{Name: "work", Role: "work", TimeDuration: 270, TimeDone: 90, Priority: 1}
		if state.TodayTasks[0] != want {
			t.Errorf("TodayTasks[0] = %+v, want %+v", state.TodayTasks[0], want)
		}
		if state.TotalPlanned != 300 || state.TotalDone != 90 || state.CompletionPct != 30 {
			t.Errorf("unexpected totals: %+v", state)
		}
	})

	t.Run("null today_tasks", func(t *testing.T) {
		client.Transport = fakeDashboard(http.StatusOK, `{"today_tasks": null, "total_planned": 0, "total_done": 0, "completion_pct": 0}`, nil)

		state, err := GetDashboardState()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.TodayTasks == nil || len(state.TodayTasks) != 0 {
			t.Errorf("expected non-nil empty list, got %#v", state.TodayTasks)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = fakeDashboard(http.StatusInternalServerError, `{"error":"fail"}`, nil)

		if _, err := GetDashboardState(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetTaskList(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		client.Transport = fakeDashboard(http.StatusOK, dashboardBodyA, nil)

		tasks, err := GetTaskList()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 2 || tasks[0].Name != "work" {
			t.Errorf("unexpected tasks: %+v", tasks)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = fakeDashboard(http.StatusInternalServerError, `{"error":"fail"}`, nil)

		if _, err := GetTaskList(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
