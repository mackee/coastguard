package main

import (
	"context"
	"fmt"
	"net/url"
	"path"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func newOAuth2ConfigAndProvider(ctx context.Context, opts *Options) (*oauth2.Config, *gooidc.Provider, error) {
	u, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "/__auth/callback")

	op, err := gooidc.NewProvider(ctx, opts.OIDCIssuer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	return &oauth2.Config{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		Endpoint:     op.Endpoint(),
		RedirectURL:  u.String(),
		Scopes:       []string{gooidc.ScopeOpenID, "email"},
	}, op, nil
}
