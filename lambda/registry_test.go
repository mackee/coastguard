package main_test

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	main "github.com/mackee/coastguard/lambda"
)

func TestNewRegistry_InvalidSession_ClearsCookieAndRedirects(t *testing.T) {
	cookieName := "coastguard_session"
	oldSecret := []byte("old-secret-key-1234567890123456")
	newSecret := []byte("new-secret-key-1234567890123456")

	// Create a server with the old secret and establish a session
	oldHandler, err := main.NewTestHandler(&main.Options{
		SessionSecret:     main.NewBase64Secret(oldSecret),
		SessionCookieName: cookieName,
	})
	if err != nil {
		t.Fatalf("failed to create old handler: %v", err)
	}
	oldServer := httptest.NewTLSServer(oldHandler)
	defer oldServer.Close()

	oldClient := oldServer.Client()
	oldClient.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}

	// Establish a session cookie via /setup-session
	resp, err := oldClient.Get("https://example.com/setup-session")
	if err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from setup-session, got %d", resp.StatusCode)
	}

	// Verify the old server accepts the cookie
	resp, err = oldClient.Get("https://example.com/ping")
	if err != nil {
		t.Fatalf("failed to request old server: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from old server, got %d", resp.StatusCode)
	}

	// Create a new server with the NEW secret key
	newHandler, err := main.NewTestHandler(&main.Options{
		SessionSecret:     main.NewBase64Secret(newSecret),
		SessionCookieName: cookieName,
	})
	if err != nil {
		t.Fatalf("failed to create new handler: %v", err)
	}
	newServer := httptest.NewTLSServer(newHandler)
	defer newServer.Close()

	newClient := newServer.Client()
	// Reuse the same cookie jar (has old session cookie)
	newClient.Jar = oldClient.Jar
	// Don't follow redirects so we can inspect the response
	newClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err = newClient.Get("https://example.com/ping")
	if err != nil {
		t.Fatalf("failed to request new server: %v", err)
	}

	// Verify redirect to /__auth/redirect
	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/__auth/redirect" {
		t.Errorf("expected Location header %q, got %q", "/__auth/redirect", loc)
	}

	// Verify cookie is cleared
	var found bool
	for _, c := range resp.Cookies() {
		if c.Name == cookieName {
			found = true
			if c.MaxAge != -1 {
				t.Errorf("expected cookie MaxAge to be -1 (delete), got %d", c.MaxAge)
			}
			if c.Value != "" {
				t.Errorf("expected cookie Value to be empty, got %q", c.Value)
			}
		}
	}
	if !found {
		t.Error("expected a Set-Cookie header to clear the session cookie, but none found")
	}
}

func TestNewRegistry_ValidSession_ReturnsRegistry(t *testing.T) {
	cookieName := "coastguard_session"
	secret := []byte("same-secret-key-1234567890123456")

	handler, err := main.NewTestHandler(&main.Options{
		SessionSecret:     main.NewBase64Secret(secret),
		SessionCookieName: cookieName,
	})
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()

	client := server.Client()
	client.Jar, err = cookiejar.New(nil)
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}

	// Establish a session cookie via /setup-session
	resp, err := client.Get("https://example.com/setup-session")
	if err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from setup-session, got %d", resp.StatusCode)
	}

	// Verify the session is valid
	resp, err = client.Get("https://example.com/ping")
	if err != nil {
		t.Fatalf("failed to request server: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
