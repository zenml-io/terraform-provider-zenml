package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

// countingTransport counts round trips, including ones that never reach a
// server.
type countingTransport struct {
	next  http.RoundTripper
	calls int32
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&t.calls, 1)
	return t.next.RoundTrip(req)
}

func TestDoRequest_DoesNotReplaySentPost(t *testing.T) {
	server, calls, _ := newDroppingServer(t, 1)
	client := NewClient(server.URL, "", "token")

	_, _, err := client.doRequest(context.Background(), "POST", "/api/v1/components", map[string]string{"name": "orchestrator"})
	if err == nil {
		t.Fatal("expected an error: a POST that was sent must not be replayed")
	}
	if got := atomic.LoadInt32(calls); got != 1 {
		t.Fatalf("expected 1 attempt, got %d", got)
	}
}

func TestDoRequest_RetriesUnsentPost(t *testing.T) {
	// A closed server refuses the connection, so the request is never written.
	server := httptest.NewServer(http.NotFoundHandler())
	serverURL := server.URL
	server.Close()

	client := NewClient(serverURL, "", "token")
	transport := &countingTransport{next: client.HTTPClient.Transport}
	client.HTTPClient.Transport = transport

	_, _, err := client.doRequest(context.Background(), "POST", "/api/v1/components", map[string]string{"name": "orchestrator"})
	if err == nil {
		t.Fatal("expected an error when the server is unreachable")
	}
	if got := atomic.LoadInt32(&transport.calls); got != maxRequestAttempts {
		t.Fatalf("expected %d attempts, got %d", maxRequestAttempts, got)
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
