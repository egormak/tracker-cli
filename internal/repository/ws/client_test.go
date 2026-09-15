package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestWSClient_ConnectAndReceive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		event := Event{
			Type:     EventTaskAdjusted,
			TaskName: "work",
			Duration: 30,
		}
		data, _ := json.Marshal(event)
		_ = conn.WriteMessage(websocket.TextMessage, data)

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	domain := strings.TrimPrefix(server.URL, "http://")
	client := NewClient("http://" + domain)
	events := client.Start()
	defer client.Close()

	select {
	case ev := <-events:
		if ev.Type != EventTaskAdjusted {
			t.Errorf("expected EventTaskAdjusted, got %v", ev.Type)
		}
		if ev.TaskName != "work" {
			t.Errorf("expected task_name work, got %v", ev.TaskName)
		}
		if ev.Duration != 30 {
			t.Errorf("expected duration 30, got %v", ev.Duration)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for WS event")
	}
}
