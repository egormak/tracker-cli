package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type EventType string

const (
	EventTaskStarted  EventType = "TASK_STARTED"
	EventTaskPaused   EventType = "TASK_PAUSED"
	EventTaskResumed  EventType = "TASK_RESUMED"
	EventTaskStopped  EventType = "TASK_STOPPED"
	EventTaskAdjusted EventType = "TASK_ADJUSTED"
	EventHeartbeatAck EventType = "HEARTBEAT_ACK"
	EventStateSync    EventType = "STATE_SYNC"
)

type Event struct {
	Type       EventType       `json:"type"`
	TaskName   string          `json:"task_name,omitempty"`
	Role       string          `json:"role,omitempty"`
	Duration   int             `json:"duration,omitempty"`
	Reason     string          `json:"reason,omitempty"`
	ServerTime int64           `json:"server_time"`
	Data       json.RawMessage `json:"data,omitempty"`
}

type Client struct {
	serverURL string
	conn      *websocket.Conn
	events    chan Event
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	closed    bool
}

func NewClient(serverDomain string) *Client {
	wsURL := convertToWSURL(serverDomain) + "/api/v1/timer/ws"
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		serverURL: wsURL,
		events:    make(chan Event, 64),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func convertToWSURL(domain string) string {
	if strings.HasPrefix(domain, "https://") {
		return "wss://" + strings.TrimPrefix(domain, "https://")
	}
	if strings.HasPrefix(domain, "http://") {
		return "ws://" + strings.TrimPrefix(domain, "http://")
	}
	return "ws://" + domain
}

func (c *Client) Start() <-chan Event {
	go c.runLoop()
	return c.events
}

func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.cancel()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.mu.Unlock()
}

func (c *Client) runLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		u, err := url.Parse(c.serverURL)
		if err != nil {
			slog.Warn("WebSocket invalid URL", "url", c.serverURL, "error", err)
			return
		}

		conn, _, err := websocket.DefaultDialer.DialContext(c.ctx, u.String(), nil)
		if err != nil {
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var event Event
			if err := json.Unmarshal(message, &event); err == nil {
				select {
				case c.events <- event:
				default:
				}
			}
		}

		c.mu.Lock()
		_ = conn.Close()
		c.conn = nil
		c.mu.Unlock()

		select {
		case <-c.ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
