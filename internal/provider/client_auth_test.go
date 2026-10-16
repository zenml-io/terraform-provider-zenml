package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newAuthServer returns a server that answers logins with loginStatus and
// loginBody, and every other request with an empty JSON object.
func newAuthServer(t *testing.T, loginStatus int, loginBody string) (*httptest.Server, *int32) {
	t.Helper()
	var logins int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == loginPath {
			atomic.AddInt32(&logins, 1)
			w.WriteHeader(loginStatus)
			_, _ = w.Write([]byte(loginBody))
			return
		}
		if r.Header.Get("Authorization") != "Bearer fresh-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	return server, &logins
}

func TestGetAPIToken_ReportsLoginFailure(t *testing.T) {
	server, _ := newAuthServer(t, http.StatusUnauthorized, `{"detail": "Invalid API key"}`)
	client := NewClient(server.URL, "bad-key", "")

	_, err := client.getAPIToken(context.Background())
	if err == nil {
		t.Fatal("expected the failed login to be reported")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Invalid API key") {
		t.Fatalf("expected the status and server message in the error, got: %v", err)
	}
	if client.APIToken != "" {
		t.Fatalf("a failed login must not cache a token, got %q", client.APIToken)
	}
}

func TestGetAPIToken_LogsInOnceForConcurrentRequests(t *testing.T) {
	server, logins := newAuthServer(t, http.StatusOK, `{"access_token": "fresh-token", "expires_in": 3600}`)
	client := NewClient(server.URL, "key", "")

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, _, err := client.doRequest(context.Background(), "GET", "/api/v1/info", nil)
			if err != nil {
				errs <- err
				return
			}
			_ = resp.Body.Close()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("request failed: %v", err)
	}
	if got := atomic.LoadInt32(logins); got != 1 {
		t.Fatalf("expected a single login, got %d", got)
	}
}

func TestGetAPIToken_ShortLivedTokenIsUsable(t *testing.T) {
	server, logins := newAuthServer(t, http.StatusOK, `{"access_token": "fresh-token", "expires_in": 60}`)
	client := NewClient(server.URL, "key", "")

	for i := 0; i < 3; i++ {
		if _, err := client.getAPIToken(context.Background()); err != nil {
			t.Fatalf("login failed: %v", err)
		}
	}
	if got := atomic.LoadInt32(logins); got != 1 {
		t.Fatalf("a token valid for 60s should be reused, got %d logins", got)
	}
	if client.APITokenExpires == nil || !client.APITokenExpires.After(time.Now()) {
		t.Fatalf("expected the token expiry to be in the future, got %v", client.APITokenExpires)
	}
}

func TestGetAPIToken_RetriesLoginAfterConnectionDrop(t *testing.T) {
	server, calls, _ := newDroppingServer(t, 1)
	client := NewClient(server.URL, "key", "")

	// The dropping server answers the retried login with {"id": "123"},
	// which has no access token.
	_, err := client.getAPIToken(context.Background())
	if got := atomic.LoadInt32(calls); got != 2 {
		t.Fatalf("expected the login to be retried once, got %d attempts", got)
	}
	if err == nil || !strings.Contains(err.Error(), "no access token") {
		t.Fatalf("expected a missing token error, got: %v", err)
	}
}

func TestIsSensitivePath(t *testing.T) {
	for path, want := range map[string]bool{
		"/api/v1/secrets":                   true,
		"/api/v1/secrets/123":               true,
		"/api/v1/service_connectors":        true,
		"/api/v1/service_connectors/verify": true,
		"/api/v1/components":                false,
		"/api/v1/stacks/123":                false,
	} {
		if got := isSensitivePath(path); got != want {
			t.Errorf("isSensitivePath(%q) = %v, want %v", path, got, want)
		}
	}
}
