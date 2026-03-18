//go:build !windows

package secret

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func socketPath() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "docker-secrets-engine", "engine.sock")
	}
	return filepath.Join(os.TempDir(), "docker-secrets-engine", "engine.sock")
}

// newHTTPClient creates a fresh HTTP client for each request.
// This avoids connection state issues with Unix sockets that can cause hangs.
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath())
			},
			DisableKeepAlives: true,
		},
	}
}
