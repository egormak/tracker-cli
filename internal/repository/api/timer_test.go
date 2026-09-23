package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestTimeDurationGet(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	tests := []struct {
		name         string
		responseJSON string
		statusCode   int
		expected     int
	}{
		{
			name:         "time_duration field present",
			responseJSON: `{"time_duration": 25}`,
			statusCode:   http.StatusOK,
			expected:     25,
		},
		{
			name:         "count field present",
			responseJSON: `{"count": 25}`,
			statusCode:   http.StatusOK,
			expected:     25,
		},
		{
			name:         "status accept with both fields",
			responseJSON: `{"status": "accept", "count": 25, "time_duration": 25}`,
			statusCode:   http.StatusOK,
			expected:     25,
		},
		{
			name:         "server error 500",
			responseJSON: `{"error": "internal error"}`,
			statusCode:   http.StatusInternalServerError,
			expected:     0,
		},
		{
			name:         "malformed json",
			responseJSON: `invalid json`,
			statusCode:   http.StatusOK,
			expected:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodGet && req.URL.Path == "/api/v1/timer/get" {
					return &http.Response{
						StatusCode: tc.statusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tc.responseJSON)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": "not found"}`)),
					Header:     make(http.Header),
				}, nil
			})

			got := TimeDurationGet()
			if got != tc.expected {
				t.Errorf("TimeDurationGet() = %d, expected %d", got, tc.expected)
			}
		})
	}
}

func TestTimeListSet(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var receivedReq timerSetRequest
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost && req.URL.Path == "/api/v1/timer/set" {
				_ = json.NewDecoder(req.Body).Decode(&receivedReq)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status":"accept","message":"Timer was set"}`)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := TimeListSet(30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedReq.Count != 30 || receivedReq.TimeDuration != 30 {
			t.Errorf("expected count=30, time_duration=30, got count=%d, time_duration=%d", receivedReq.Count, receivedReq.TimeDuration)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"server error"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := TimeListSet(30)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestTimeDurationDel(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var receivedReq timerSetRequest
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost && req.URL.Path == "/api/v1/timer/del" {
				_ = json.NewDecoder(req.Body).Decode(&receivedReq)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status":"accept","message":"Timer was deleted"}`)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := TimeDurationDel(15)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedReq.Count != 15 || receivedReq.TimeDuration != 15 {
			t.Errorf("expected count=15, time_duration=15, got count=%d, time_duration=%d", receivedReq.Count, receivedReq.TimeDuration)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"server error"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := TimeDurationDel(15)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestSetGlobalTime(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var receivedReq timerGlobalSetRequest
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost && req.URL.Path == "/api/v1/manage/timer/global" {
				_ = json.NewDecoder(req.Body).Decode(&receivedReq)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status":"accept","message":"Timer set"}`)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := SetGlobalTime(60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedReq.TimeScheduler != 60 {
			t.Errorf("expected time_scheduler=60, got %d", receivedReq.TimeScheduler)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"server error"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := SetGlobalTime(60)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetGlobalTime(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && req.URL.Path == "/api/v1/manage/timer/global" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"timer_global": 45}`)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		})

		got, err := GetGlobalTime()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 45 {
			t.Errorf("expected 45, got %d", got)
		}
	})

	t.Run("server error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"server error"}`)),
				Header:     make(http.Header),
			}, nil
		})

		_, err := GetGlobalTime()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("decode error", func(t *testing.T) {
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`invalid json`)),
				Header:     make(http.Header),
			}, nil
		})

		_, err := GetGlobalTime()
		if err == nil {
			t.Fatal("expected decode error, got nil")
		}
	})
}
