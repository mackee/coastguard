package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"text/template"

	"github.com/mackee/tanukirpc"
	"github.com/mackee/tanukirpc/auth/oidc"
)

//go:embed redirect.html
var fs embed.FS
var tmpl = template.Must(template.New("redirect.html").ParseFS(fs, "redirect.html"))

func newHandler(opts *Options) (http.Handler, error) {
	rf, err := newRegistryFactory(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry factory: %w", err)
	}

	r := tanukirpc.NewRouter(&registry{},
		tanukirpc.WithContextFactory(
			tanukirpc.NewContextHookFactory(rf.NewRegistry),
		),
	)
	oauth2Config, oidcProvider, err := newOAuth2ConfigAndProvider(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 config: %w", err)
	}
	oidcOptions := []oidc.HandlersOption[*registry]{
		oidc.WithDefaultReferrer[*registry]("/"),
		oidc.WithReferrerBaseURL[*registry](opts.BaseURL),
		oidc.WithUnauthorizedRedirect[*registry]("/"),
		oidc.WithSuccessBehavior(func(ctx tanukirpc.Context[*registry], input *oidc.SuccessBehaviorInput) error {
			resp := ctx.Response()
			SetCookiesPresign(resp, opts)
			http.SetCookie(resp, &http.Cookie{
				Name:     opts.SessionCookieName,
				Value:    "",
				Quoted:   false,
				Path:     "",
				Domain:   "",
				MaxAge:   -1,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			return nil
		}),
	}
	if len(opts.AllowedDomains) > 0 {
		oidcOptions = append(oidcOptions, oidc.WithAllowedDomains[*registry](opts.AllowedDomains...))
	}

	oidcAuth := oidc.NewHandlers(
		oauth2Config,
		oidcProvider,
		oidcOptions...,
	)

	r.Route("/__auth", func(router *tanukirpc.Router[*registry]) {
		router.Get("/redirect", tanukirpc.NewHandler(
			func(ctx tanukirpc.Context[*registry], _ struct{}) (_resp struct{}, err error) {
				opts := ctx.Registry().options
				bypassIPs := opts.BypassIPs
				address := ctx.Request().Header.Get("CloudFront-Viewer-Address")
				if isBypassIP(address, bypassIPs) {
					SetCookiesPresign(ctx.Response(), opts)
					return struct{}{}, redirectToReferrer(ctx)
				}

				return oidcAuth.Redirect(ctx, struct{}{})
			},
		))
		router.Get("/callback", tanukirpc.NewHandler(oidcAuth.Callback))
		router.Get("/logout", tanukirpc.NewHandler(logoutHandler))
	})
	r.Get("/__auth/unauthorized", tanukirpc.NewHandler(unauthorizedHandler))

	return r, nil
}

func logoutHandler(ctx tanukirpc.Context[*registry], _ struct{}) (_resp struct{}, err error) {
	RemovePresignedCookie(ctx.Response())

	return struct{}{}, nil
}

type templateArgs struct {
	RedirectTo string
}

func unauthorizedHandler(ctx tanukirpc.Context[*registry], _ struct{}) (_resp struct{}, err error) {
	resp := ctx.Response()
	// Check if the request is from a bypass IP
	resp.Header().Set("Content-Type", "text/html")
	resp.WriteHeader(http.StatusUnauthorized)
	if err := tmpl.ExecuteTemplate(resp, "redirect.html", templateArgs{RedirectTo: "/__auth/redirect"}); err != nil {
		return struct{}{}, fmt.Errorf("failed to execute template: %w", err)
	}

	return struct{}{}, nil
}

func isBypassIP(address string, bypassIPs []string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	return slices.Contains(bypassIPs, host)
}

func redirectToReferrer(ctx tanukirpc.Context[*registry]) error {
	req := ctx.Request()
	referrer := req.Referer()
	if referrer == "" {
		return tanukirpc.ErrorRedirectTo(http.StatusFound, "/")
	}
	if strings.HasPrefix(referrer, "/") {
		return tanukirpc.ErrorRedirectTo(http.StatusFound, referrer)
	}
	if trimmed := strings.TrimPrefix(referrer, ctx.Registry().options.BaseURL); trimmed != referrer {
		if trimmed == "" {
			trimmed = "/"
		}
		if !strings.HasPrefix(trimmed, "/") {
			trimmed = "/" + trimmed
		}
		return tanukirpc.ErrorRedirectTo(http.StatusFound, trimmed)
	}

	return tanukirpc.ErrorRedirectTo(http.StatusFound, "/")
}
