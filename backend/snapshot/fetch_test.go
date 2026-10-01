package snapshot

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchSnapshot(t *testing.T) {
	const body = `{"state":{"schemaVersion":1}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected request method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = io.WriteString(w, body)
	}))
	defer server.Close()

	got, err := fetchSnapshot(server.Client(), server.URL)
	if err != nil || string(got) != body {
		t.Fatalf("fetchSnapshot() = %q, %v; want %q", got, err, body)
	}
}

func TestFetchSnapshotRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		contentType string
		body        string
		wantError   string
	}{
		{"HTTP error", http.StatusServiceUnavailable, "application/json", `{}`, "HTTP 503"},
		{"missing content type", http.StatusOK, "", `{}`, "expected application/json"},
		{"HTML", http.StatusOK, "text/html", `<html></html>`, "expected application/json"},
		{"malformed media type", http.StatusOK, "application/json; bad", `{}`, "expected application/json"},
		{"oversized body", http.StatusOK, "application/json", strings.Repeat("x", maxSnapshotBytes+1), "too large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer server.Close()

			got, err := fetchSnapshot(server.Client(), server.URL)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) || got != nil {
				t.Fatalf("fetchSnapshot() = %q, %v; want nil and error containing %q", got, err, tt.wantError)
			}
		})
	}
}

func TestFetchSnapshotDoesNotTrustContentLength(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{"Content-Type": {"application/json"}},
			Body:          io.NopCloser(strings.NewReader(strings.Repeat("x", maxSnapshotBytes+1))),
			ContentLength: -1,
			Request:       req,
		}, nil
	})}
	got, err := fetchSnapshot(client, "https://example.test/snapshot.php")
	var sizeErr *http.MaxBytesError
	if !errors.As(err, &sizeErr) || got != nil {
		t.Fatalf("fetchSnapshot() = %q, %v; want size error and no data", got, err)
	}
}

func TestFetchSnapshotHonorsClientTimeout(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Millisecond, Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}
	got, err := fetchSnapshot(client, "https://example.test/snapshot.php")
	if !errors.Is(err, context.DeadlineExceeded) || got != nil || client.Timeout != 10*time.Millisecond {
		t.Fatalf("fetchSnapshot() = %q, %v; want deadline exceeded and no data", got, err)
	}
}

func TestFetchSnapshotRejectsBadRequest(t *testing.T) {
	client := &http.Client{}
	for _, tt := range []struct {
		client *http.Client
		url    string
	}{
		{nil, "https://example.test/snapshot.php"},
		{client, "://bad-url"},
	} {
		got, err := fetchSnapshot(tt.client, tt.url)
		if err == nil || got != nil {
			t.Fatalf("fetchSnapshot(%q) = %q, %v; want error and no data", tt.url, got, err)
		}
	}
}
