package updater

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// requireSecureURL rejects any base URL that is not HTTPS, unless it points
// at loopback (127.0.0.1, ::1, or localhost), so local development against
// http://127.0.0.1:3001 keeps working. Without this check, a plain `go
// build` or one bad Taskfile edit — the production value is injected only
// via LDFLAGS — silently produces a client that fetches and executes a
// binary over cleartext HTTP.
func requireSecureURL(rawURL string) error {
	u, err := url.Parse(rawURL)

	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	if u.Scheme == "https" {
		return nil
	}

	if isLoopbackHost(u.Hostname()) {
		return nil
	}

	return fmt.Errorf("refusing non-HTTPS URL %q", rawURL)
}

// isLoopbackHost reports whether host (as returned by url.URL.Hostname, so
// already stripped of port and brackets) names the local machine.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}
