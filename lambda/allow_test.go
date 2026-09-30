package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/mackee/tanukirpc/auth/oidc"
)

const testIssuer = "https://issuer.example.com"

func newTestIDToken(t *testing.T, claims map[string]any) *gooidc.IDToken {
	t.Helper()
	claims["iss"] = testIssuer
	claims["sub"] = "subject"
	claims["aud"] = "client"
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	header, err := json.Marshal(map[string]string{"alg": "RS256"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding.EncodeToString
	raw := enc(header) + "." + enc(payload) + "." + enc([]byte("signature"))

	verifier := gooidc.NewVerifier(testIssuer, &gooidc.StaticKeySet{}, &gooidc.Config{
		ClientID:                   "client",
		InsecureSkipSignatureCheck: true,
	})
	idToken, err := verifier.Verify(context.Background(), raw)
	if err != nil {
		t.Fatal(err)
	}
	return idToken
}

func TestNewAllowFunc_NoRestriction(t *testing.T) {
	if _, ok := newAllowFunc(&Options{}); ok {
		t.Fatal("expected no allow func when neither AllowedDomains nor AllowedEmails is set")
	}
}

func TestNewAllowFunc(t *testing.T) {
	workspaceUser := map[string]any{"hd": "example.com", "email": "alice@example.com", "email_verified": true}
	guestUser := map[string]any{"email": "Guest@Gmail.com", "email_verified": true}
	unverifiedGuest := map[string]any{"email": "guest@gmail.com", "email_verified": false}
	stranger := map[string]any{"email": "stranger@gmail.com", "email_verified": true}

	tests := []struct {
		name    string
		opts    *Options
		claims  map[string]any
		allowed bool
	}{
		{"domains only: matching hd", &Options{AllowedDomains: []string{"example.com"}}, workspaceUser, true},
		{"domains only: no hd", &Options{AllowedDomains: []string{"example.com"}}, guestUser, false},
		{"emails only: matching email", &Options{AllowedEmails: []string{"guest@gmail.com"}}, guestUser, true},
		{"emails only: workspace user not listed", &Options{AllowedEmails: []string{"guest@gmail.com"}}, workspaceUser, false},
		{"emails only: unverified email", &Options{AllowedEmails: []string{"guest@gmail.com"}}, unverifiedGuest, false},
		{"both: matching hd", &Options{AllowedDomains: []string{"example.com"}, AllowedEmails: []string{"guest@gmail.com"}}, workspaceUser, true},
		{"both: matching email", &Options{AllowedDomains: []string{"example.com"}, AllowedEmails: []string{"guest@gmail.com"}}, guestUser, true},
		{"both: matching neither", &Options{AllowedDomains: []string{"example.com"}, AllowedEmails: []string{"guest@gmail.com"}}, stranger, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := newAllowFunc(tt.opts)
			if !ok {
				t.Fatal("expected allow func")
			}
			err := fn(nil, newTestIDToken(t, tt.claims))
			if tt.allowed {
				if err != nil {
					t.Fatalf("expected allowed, got %v", err)
				}
				return
			}
			if !errors.Is(err, oidc.ErrNotAllowed) {
				t.Fatalf("expected ErrNotAllowed, got %v", err)
			}
		})
	}
}
