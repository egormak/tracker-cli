package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"tracker_cli/internal/domain/entity"
)

func TestGetRampStatus(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var called bool
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == "GET" && req.URL.Path == "/api/v1/ramp/status" {
				called = true
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
							"excluded_tasks": ["video", "movies", "games", "telegram"],
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
		})

		status, err := GetRampStatus()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected GET /api/v1/ramp/status to be called")
		}
		if status.CurrentStep != 3 || status.CapMinutes != 25 || status.TodayFocusMinutes != 6 {
			t.Errorf("unexpected status parsed: %+v", status)
		}
		if len(status.Config.EnabledRoles) != 2 || status.Config.DefaultRestFallback != 15 {
			t.Errorf("unexpected config inside status: %+v", status.Config)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"internal error"}`)),
				Header:     make(http.Header),
			}, nil
		})

		_, err := GetRampStatus()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestResetRamp(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var called bool
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == "POST" && req.URL.Path == "/api/v1/ramp/reset" {
				called = true
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
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		})

		status, err := ResetRamp()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected POST /api/v1/ramp/reset to be called")
		}
		if status.CurrentStep != 1 {
			t.Errorf("expected CurrentStep 1, got %d", status.CurrentStep)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"fail"}`)),
				Header:     make(http.Header),
			}, nil
		})

		_, err := ResetRamp()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAdvanceRamp(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	var called bool
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == "POST" && req.URL.Path == "/api/v1/ramp/advance" {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"current_step": 4,
					"cap_minutes": 25,
					"is_capped": false,
					"today_focus_minutes": 6,
					"date": "23 September 2026",
					"config": {
						"cap_minutes": 25,
						"enabled_roles": ["work"],
						"enabled_tasks": [],
						"excluded_tasks": [],
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
	})

	status, err := AdvanceRamp()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected POST /api/v1/ramp/advance to be called")
	}
	if status.CurrentStep != 4 {
		t.Errorf("expected CurrentStep 4, got %d", status.CurrentStep)
	}
}

func TestGetRampConfig(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	var called bool
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == "GET" && req.URL.Path == "/api/v1/ramp/config" {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"cap_minutes": 20,
					"enabled_roles": ["work", "learn"],
					"enabled_tasks": ["home_task"],
					"excluded_tasks": ["video", "movies"],
					"default_rest_fallback": 12
				}`)),
				Header: make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
			Header:     make(http.Header),
		}, nil
	})

	cfg, err := GetRampConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected GET /api/v1/ramp/config to be called")
	}
	if cfg.CapMinutes != 20 || cfg.DefaultRestFallback != 12 {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestUpdateRampConfig(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	var called bool
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == "PUT" && req.URL.Path == "/api/v1/ramp/config" {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"current_step": 3,
					"cap_minutes": 30,
					"is_capped": false,
					"today_focus_minutes": 6,
					"date": "23 September 2026",
					"config": {
						"cap_minutes": 30,
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
	})

	cfg := entity.RampConfig{
		CapMinutes:          30,
		EnabledRoles:        []string{"work", "learn"},
		EnabledTasks:        []string{"home_task"},
		ExcludedTasks:       []string{"video"},
		DefaultRestFallback: 15,
	}

	status, err := UpdateRampConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected PUT /api/v1/ramp/config to be called")
	}
	if status.CapMinutes != 30 {
		t.Errorf("expected CapMinutes 30, got %d", status.CapMinutes)
	}
}
