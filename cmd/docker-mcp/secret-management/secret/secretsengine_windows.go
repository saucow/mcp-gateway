//go:build windows

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
// On Windows, the secrets engine Unix socket (engine.sock) is an AF_UNIX
// reparse point into WSL2. This works for non-MSIX processes but fails for
// MSIX-sandboxed processes (e.g. Claude Desktop) with WSAEINVAL.
// When this fails, BuildSecretsURIs() in secrets_uri.go falls back to
// generating se:// URIs that Docker Desktop resolves at container runtime.
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
