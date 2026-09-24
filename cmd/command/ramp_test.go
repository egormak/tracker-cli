package command

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"tracker_cli/internal/repository/api"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRampCommand_Configuration(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"ramp"})
	if err != nil {
		t.Fatalf("expected to find 'ramp' command: %v", err)
	}
	if cmd == nil || cmd.Name() != "ramp" {
		t.Fatal("expected 'ramp' command to be registered on rootCmd")
	}

	subCommands := []string{"status", "reset", "set-cap"}
	for _, sub := range subCommands {
		c, _, err := rootCmd.Find([]string{"ramp", sub})
		if err != nil || c == nil || c.Name() != sub {
			t.Errorf("expected sub-command 'ramp %s' to be registered", sub)
		}
	}
}

func TestRampCommand_Execution(t *testing.T) {
	cleanup := api.SetClientTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == "GET" && req.URL.Path == "/api/v1/ramp/status":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"current_step": 3,
					"cap_minutes": 25,
					"is_capped": false,
					"today_focus_minutes": 6,
					"date": "23 September 2026",
					"config": {
						"cap_minutes": 25,
						"enabled_roles": ["work", "learn"],
						"enabled_tasks": ["home_task"],
						"excluded_tasks": ["video"],
						"default_rest_fallback": 15
					}
				}`)),
				Header: make(http.Header),
			}, nil
		case req.Method == "POST" && req.URL.Path == "/api/v1/ramp/reset":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"current_step": 1,
					"cap_minutes": 25,
					"is_capped": false,
					"today_focus_minutes": 6,
					"date": "23 September 2026",
					"config": {
						"cap_minutes": 25,
						"enabled_roles": ["work", "learn"],
						"enabled_tasks": ["home_task"],
						"excluded_tasks": ["video"],
						"default_rest_fallback": 15
					}
				}`)),
				Header: make(http.Header),
			}, nil
		case req.Method == "PUT" && req.URL.Path == "/api/v1/ramp/config":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"current_step": 3,
					"cap_minutes": 20,
					"is_capped": false,
					"today_focus_minutes": 6,
					"date": "23 September 2026",
					"config": {
						"cap_minutes": 20,
						"enabled_roles": ["work", "learn"],
						"enabled_tasks": ["home_task"],
						"excluded_tasks": ["video"],
						"default_rest_fallback": 15
					}
				}`)),
				Header: make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
			Header:     make(http.Header),
		}, nil
	}))
	defer cleanup()

	t.Run("ramp status output", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rampCmd.SetOut(buf)
		rampCmd.SetErr(buf)

		err := rampCmd.RunE(rampCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Warm-Up Ramp Ladder") {
			t.Errorf("expected header in output, got: %s", out)
		}
		if !strings.Contains(out, "[3 мин]") {
			t.Errorf("expected current step [3 мин] in output, got: %s", out)
		}
	})

	t.Run("ramp reset output", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rampResetCmd.SetOut(buf)
		rampResetCmd.SetErr(buf)

		err := rampResetCmd.RunE(rampResetCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Warm-Up Ramp reset to 1 minute") {
			t.Errorf("expected reset message, got: %s", out)
		}
	})

	t.Run("ramp set-cap validation: < 5 fails", func(t *testing.T) {
		err := rampSetCapCmd.RunE(rampSetCapCmd, []string{"3"})
		if err == nil {
			t.Fatal("expected error for cap < 5, got nil")
		}
	})

	t.Run("ramp set-cap success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rampSetCapCmd.SetOut(buf)
		rampSetCapCmd.SetErr(buf)

		err := rampSetCapCmd.RunE(rampSetCapCmd, []string{"20"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Warm-Up Ramp cap set to 20 minutes") {
			t.Errorf("expected set-cap message, got: %s", out)
		}
	})
}
