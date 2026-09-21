package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResetRest(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	t.Run("success", func(t *testing.T) {
		var called bool
		client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == "POST" && req.URL.Path == "/api/v1/rest/reset" {
				called = true
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status":"accept","message":"Rest was reset"}`)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := ResetRest()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !called {
			t.Error("expected POST /api/v1/rest/reset to be called")
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

		err := ResetRest()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
