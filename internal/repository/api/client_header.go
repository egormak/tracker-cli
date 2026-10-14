package api

import (
	"net/http"
	"net/url"
	"sync"
	"tracker_cli/config"
)

// ClientHeader tells tracker-server which client a request comes from; the server stores
// it on task records and in the weekly usage counters.
const (
	ClientHeader = "X-Tracker-Client"
	ClientName   = "cli"
)

type clientHeaderTransport struct {
	base http.RoundTripper
	host string
}

func (t clientHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == t.host {
		req = req.Clone(req.Context())
		req.Header.Set(ClientHeader, ClientName)
	}
	return t.base.RoundTrip(req)
}

var installClientHeader sync.Once

// InstallClientHeader wraps http.DefaultTransport so every request to the tracker server
// carries the client name. It goes through the default transport because several packages
// build their own http.Client instead of using sendRequest.
func InstallClientHeader() {
	installClientHeader.Do(func() {
		host := ""
		if u, err := url.Parse(config.TrackerDomain); err == nil {
			host = u.Host
		}
		http.DefaultTransport = clientHeaderTransport{base: http.DefaultTransport, host: host}
	})
}
