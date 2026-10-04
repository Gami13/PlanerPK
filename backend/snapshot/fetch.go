package snapshot

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"
)

const (
	maxSnapshotBytes = 10 << 20 // The observed response is about 1.5 MiB.
	snapshotTimeout  = 30 * time.Second
)

// Fetch retrieves a bounded JSON response without decoding or publishing it.
// It caps the request timeout without modifying the supplied client.
func Fetch(client *http.Client, url string) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("fetch snapshot: nil HTTP client")
	}

	requestClient := *client
	if requestClient.Timeout == 0 || requestClient.Timeout > snapshotTimeout {
		requestClient.Timeout = snapshotTimeout
	}

	resp, err := requestClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch snapshot: HTTP %d", resp.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return nil, fmt.Errorf("fetch snapshot: expected application/json, got %q", resp.Header.Get("Content-Type"))
	}
	data, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, maxSnapshotBytes))
	if err != nil {
		return nil, fmt.Errorf("read snapshot response: %w", err)
	}
	return data, nil
}
