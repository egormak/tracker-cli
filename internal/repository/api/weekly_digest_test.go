package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestGetWeeklyDigest(t *testing.T) {
	oldTransport := client.Transport
	defer func() { client.Transport = oldTransport }()

	var gotQuery string
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" || req.URL.Path != "/api/v1/stats/weekly/digest" {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(`{}`)), Header: make(http.Header)}, nil
		}
		gotQuery = req.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(bytes.NewBufferString(`{
				"week": "previous",
				"week_start_date": "5 October 2026",
				"week_end_date": "11 October 2026",
				"total_planned": 310,
				"total_done": 320,
				"roles": [{"role": "work", "planned": 200, "done": 200, "share_percent": 62}],
				"strict": [], "free": [], "free_done": 0,
				"over_plan": [{"name": "video", "role": "rest", "planned": 10, "done": 90, "delta": 80}],
				"untouched": [], "unscheduled": [],
				"days": [{"day": "monday", "date": "5 October 2026", "total_done": 150}],
				"busiest_day": {"day": "monday", "date": "5 October 2026", "total_done": 150}
			}`)),
			Header: make(http.Header),
		}, nil
	})

	d, err := GetWeeklyDigest("previous")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "week=previous" {
		t.Errorf("expected query week=previous, got %q", gotQuery)
	}
	if d.TotalDone != 320 || len(d.Roles) != 1 || d.OverPlan[0].Delta != 80 {
		t.Errorf("unexpected digest: %+v", d)
	}
	if d.BusiestDay == nil || d.BusiestDay.Day != "monday" {
		t.Errorf("expected busiest day monday, got %+v", d.BusiestDay)
	}

	if _, err := GetWeeklyDigest(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("expected no query for the default week, got %q", gotQuery)
	}
}
