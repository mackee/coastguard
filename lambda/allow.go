package main

import (
	"github.com/mackee/tanukirpc/auth/oidc"
)

// newAllowFunc returns an AllowFunc that allows users matching any of AllowedDomains or AllowedEmails.
// It returns false if neither is set, which means all authenticated users are allowed.
func newAllowFunc(opts *Options) (oidc.AllowFunc[*registry], bool) {
	var fns []oidc.AllowFunc[*registry]
	if len(opts.AllowedDomains) > 0 {
		fns = append(fns, oidc.AllowDomains[*registry](opts.AllowedDomains...))
	}
	if len(opts.AllowedEmails) > 0 {
		fns = append(fns, oidc.AllowEmails[*registry](opts.AllowedEmails...))
	}
	if len(fns) == 0 {
		return nil, false
	}
	return oidc.AllowAnyOf(fns...), true
}
