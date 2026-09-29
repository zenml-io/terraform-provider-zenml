package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
)

// newDroppingServer returns a server that reads the full request and then
// closes the connection without a response for the first `drops` requests.
// From the client's side the request was sent and the reply is EOF, so it
// cannot tell whether the server processed it.
func newDroppingServer(t *testing.T, drops int32) (*httptest.Server, *int32, *[]string) {
	t.Helper()
	var calls int32
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
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
	return server, &calls, &bodies
}

// processedThenResetTransport models a request being consumed before the client
// receives a connection reset instead of the response.
type processedThenResetTransport struct {
	processed int32
}

func (t *processedThenResetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
	}
	atomic.AddInt32(&t.processed, 1)
	return nil, syscall.ECONNRESET
}

func TestDoRequest_DoesNotReplayPostAfterAmbiguousWriteError(t *testing.T) {
	transport := &processedThenResetTransport{}
	client := NewClient("http://localhost", "", "token")
	client.HTTPClient.Transport = transport

	_, _, err := client.doRequest(context.Background(), "POST", "/api/v1/components", map[string]string{"name": "orchestrator"})
	if err == nil {
		t.Fatal("expected the connection reset to be returned")
	}
	if !strings.Contains(err.Error(), syscall.ECONNRESET.Error()) {
		t.Fatalf("expected connection reset error, got %v", err)
	}
	if got := atomic.LoadInt32(&transport.processed); got != 1 {
		t.Fatalf("POST processed %d times; want 1", got)
	}
}

func TestDoRequest_RetriesIdempotentRequest(t *testing.T) {
	server, calls, bodies := newDroppingServer(t, 1)
	client := NewClient(server.URL, "", "token")

	resp, status, err := client.doRequest(context.Background(), "PUT", "/api/v1/components/123", map[string]string{"name": "orchestrator"})
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
	for i, body := range *bodies {
		if body != `{"name":"orchestrator"}` {
			t.Fatalf("attempt %d sent unexpected body %q", i+1, body)
		}
	}
}

func TestDoRequest_GivesUpAfterMaxAttempts(t *testing.T) {
	server, calls, _ := newDroppingServer(t, maxRequestAttempts)
	client := NewClient(server.URL, "", "token")

	_, _, err := client.doRequest(context.Background(), "GET", "/api/v1/components/123", nil)
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

	_, status, err := client.doRequest(context.Background(), "GET", "/api/v1/components/123", nil)
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
