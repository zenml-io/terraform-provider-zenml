package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newDroppingServer returns a server that closes the connection without a
// response for the first `drops` requests, mimicking a server that closes an
// idle keep-alive connection while a request is being sent on it.
func newDroppingServer(t *testing.T, drops int32, bodies *[]string) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		body, _ := io.ReadAll(r.Body)
		if bodies != nil {
			*bodies = append(*bodies, string(body))
		}
		if n <= drops {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("hijack failed: %v", err)
				return
			}
			conn.Close()
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": "123"}`))
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func TestDoRequest_RetriesDroppedConnection(t *testing.T) {
	var bodies []string
	server, calls := newDroppingServer(t, 1, &bodies)
	client := NewClient(server.URL, "", "token")

	resp, status, err := client.doRequest(context.Background(), "POST", "/api/v1/components", map[string]string{"name": "orchestrator"})
	if err != nil {
		t.Fatalf("expected request to succeed after retry, got: %v", err)
	}
	defer resp.Body.Close()

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if got := atomic.LoadInt32(calls); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
	// The request body must be replayed on the retry.
	for i, body := range bodies {
		if body != `{"name":"orchestrator"}` {
			t.Fatalf("attempt %d sent unexpected body %q", i+1, body)
		}
	}
}

func TestDoRequest_GivesUpAfterMaxAttempts(t *testing.T) {
	server, calls := newDroppingServer(t, maxRequestAttempts, nil)
	client := NewClient(server.URL, "", "token")

	_, _, err := client.doRequest(context.Background(), "POST", "/api/v1/components", map[string]string{"name": "orchestrator"})
	if err == nil {
		t.Fatal("expected an error after exhausting retries")
	}
	if got := atomic.LoadInt32(calls); got != maxRequestAttempts {
		t.Fatalf("expected %d attempts, got %d", maxRequestAttempts, got)
	}
}

func TestDoRequest_DoesNotRetryHTTPErrors(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client := NewClient(server.URL, "", "token")

	_, status, err := client.doRequest(context.Background(), "POST", "/api/v1/components", map[string]string{"name": "orchestrator"})
	if err == nil {
		t.Fatal("expected an error for a 500 response")
	}
	if status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", status)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 attempt, got %d", got)
	}
}

func TestNewClient_IdleConnTimeoutBelowServerKeepAlive(t *testing.T) {
	transport, ok := NewClient("http://localhost", "", "token").HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected an *http.Transport")
	}
	// uvicorn closes idle keep-alive connections after 5 seconds by default.
	if transport.IdleConnTimeout <= 0 || transport.IdleConnTimeout.Seconds() >= 5 {
		t.Fatalf("idle connection timeout %s must be below the server keep-alive timeout", transport.IdleConnTimeout)
	}
}
