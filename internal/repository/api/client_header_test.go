package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHeaderTransport_OnlyForTrackerHost(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Get(ClientHeader))
	}))
	defer srv.Close()

	send := func(host string) {
		t.Helper()
		client := http.Client{Transport: clientHeaderTransport{base: http.DefaultTransport, host: host}}
		req, err := http.NewRequest("GET", srv.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if req.Header.Get(ClientHeader) != "" {
			t.Errorf("the caller's request must not be modified")
		}
	}

	send(srv.Listener.Addr().String()) // the tracker server
	send("tracker.example:8080")       // any other host, e.g. the Telegram API

	if len(got) != 2 || got[0] != ClientName || got[1] != "" {
		t.Errorf("headers seen by the server = %q, want [%q \"\"]", got, ClientName)
	}
}
