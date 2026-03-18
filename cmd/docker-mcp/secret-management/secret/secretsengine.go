package secret

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrSecretNotFound = errors.New("secret not found")

type Envelope struct {
	ID       string            `json:"id"`
	Value    []byte            `json:"value"`
	Provider string            `json:"provider"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func GetSecrets(ctx context.Context) ([]Envelope, error) {
	pattern := `{"pattern": "docker/mcp/**"}`

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/resolver.v1.ResolverService/GetSecrets", bytes.NewReader([]byte(pattern)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := newHTTPClient()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// No secrets found
	if resp.StatusCode == http.StatusNotFound {
		return []Envelope{}, nil
	}

	var secrets map[string][]Envelope
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets response: %w", err)
	}

	return secrets["envelopes"], nil
}

// GetSecret retrieves a single secret by its full key (e.g., "docker/mcp/oauth/github").
// Returns ErrSecretNotFound if the secret does not exist.
func GetSecret(ctx context.Context, key string) (*Envelope, error) {
	pattern := fmt.Sprintf(`{"pattern": "%s"}`, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://localhost/resolver.v1.ResolverService/GetSecrets",
		bytes.NewReader([]byte(pattern)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := newHTTPClient()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSecretNotFound
	}

	var secrets map[string][]Envelope
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret response: %w", err)
	}

	envelopes := secrets["envelopes"]
	if len(envelopes) == 0 {
		return nil, ErrSecretNotFound
	}

	return &envelopes[0], nil
}
