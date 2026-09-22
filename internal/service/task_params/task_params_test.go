package task_params

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"tracker_cli/internal/domain/entity"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetTaskParams_Success(t *testing.T) {
	oldTransport := httpClient.Transport
	defer func() { httpClient.Transport = oldTransport }()

	httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/api/v1/task/params" {
			t.Errorf("expected path /api/v1/task/params, got %s", req.URL.Path)
		}
		if req.URL.Query().Get("task_name") != "coding" {
			t.Errorf("expected task_name coding, got %s", req.URL.Query().Get("task_name"))
		}

		resp := entity.TaskParams{
			Name:     "coding",
			Time:     60,
			Priority: 1,
		}
		bodyBytes, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
			Header:     make(http.Header),
		}, nil
	})

	params := GetTaskParams("coding")
	if params.Name != "coding" || params.Time != 60 || params.Priority != 1 {
		t.Errorf("unexpected params: %+v", params)
	}
}

func TestGetTaskParams_NotFound(t *testing.T) {
	oldTransport := httpClient.Transport
	defer func() { httpClient.Transport = oldTransport }()

	httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString(`{"status":"Task Not Found"}`)),
			Header:     make(http.Header),
		}, nil
	})

	params := GetTaskParams("unknown")
	if params != (TaskParams{}) {
		t.Errorf("expected empty TaskParams on 404, got %+v", params)
	}
}

func TestSetTaskParams_Success(t *testing.T) {
	oldTransport := httpClient.Transport
	defer func() { httpClient.Transport = oldTransport }()

	var receivedBody entity.TaskParams
	httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/api/v1/task/params" {
			t.Errorf("expected path /api/v1/task/params, got %s", req.URL.Path)
		}
		if req.Method != "POST" {
			t.Errorf("expected POST, got %s", req.Method)
		}
		json.NewDecoder(req.Body).Decode(&receivedBody)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"status":"success"}`)),
			Header:     make(http.Header),
		}, nil
	})

	SetTaskParams("review", 30, 2)
	if receivedBody.Name != "review" || receivedBody.Time != 30 || receivedBody.Priority != 2 {
		t.Errorf("unexpected body received: %+v", receivedBody)
	}
}
