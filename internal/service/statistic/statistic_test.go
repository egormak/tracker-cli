package statistic

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"tracker_cli/internal/repository/api"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

const dashboardBodyA = `{"running_task": null, "today_tasks": [{"name": "work", "role": "work", "time_duration": 270, "time_done": 90, "priority": 1}, {"name": "english", "role": "learn", "time_duration": 30, "time_done": 0, "priority": 2}], "evening_focus": null, "total_planned": 300, "total_done": 90, "completion_pct": 30}`

// runShowAfterTask serves body as the dashboard state and returns the log output.
func runShowAfterTask(t *testing.T, body, task string) string {
	t.Helper()

	cleanup := api.SetClientTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}))
	defer cleanup()

	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	ShowAfterTask(task)
	return buf.String()
}

func TestShowAfterTask(t *testing.T) {
	t.Run("task in today's list", func(t *testing.T) {
		out := runShowAfterTask(t, dashboardBodyA, "work")
		for _, want := range []string{"task=work", "time_duration=90", "left=180", "percent=30", "Time Done"} {
			if !strings.Contains(out, want) {
				t.Errorf("output lacks %q:\n%s", want, out)
			}
		}
	})

	t.Run("no plan for today", func(t *testing.T) {
		out := runShowAfterTask(t, `{"today_tasks": [], "total_planned": 0, "total_done": 0, "completion_pct": 0}`, "work")
		if !strings.Contains(out, "Time Limit for Today is not Set") {
			t.Errorf("output lacks the no-limit line:\n%s", out)
		}
		if strings.Contains(out, "Percent Done") {
			t.Errorf("unexpected Percent Done:\n%s", out)
		}
	})

	t.Run("task outside today's list", func(t *testing.T) {
		out := runShowAfterTask(t, dashboardBodyA, "missing")
		if !strings.Contains(out, "time_duration=0") {
			t.Errorf("output lacks time_duration=0:\n%s", out)
		}
		if strings.Contains(out, "left=") {
			t.Errorf("unexpected left=:\n%s", out)
		}
	})
}

func TestStatCompletionPercentage(t *testing.T) {
	if got := statCompletionPercentage(90, 270); got != 33.33 {
		t.Errorf("statCompletionPercentage(90, 270) = %v, want 33.33", got)
	}
}
