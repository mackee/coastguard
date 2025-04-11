package main

import (
	"fmt"
	"net/http"

	gsessions "github.com/gorilla/sessions"
	"github.com/mackee/tanukirpc/sessions"
	"github.com/mackee/tanukirpc/sessions/gorilla"
)

type registryFactory struct {
	options      *Options
	sessionStore sessions.Store
}

func newRegistryFactory(opts *Options) (*registryFactory, error) {
	gss := gsessions.NewCookieStore(opts.SessionSecret.Bytes())
	tss, err := gorilla.NewStore(gss, gorilla.WithSessionName(opts.SessionCookieName))
	if err != nil {
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}

	return &registryFactory{
		options:      opts,
		sessionStore: tss,
	}, nil
}

type registry struct {
	options         *Options
	sessionAccessor sessions.Accessor
}

func (r *registryFactory) NewRegistry(w http.ResponseWriter, req *http.Request) (*registry, error) {
	session, err := r.sessionStore.GetAccessor(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &registry{
		options:         r.options,
		sessionAccessor: session,
	}, nil
}

func (r *registry) Session() sessions.Accessor {
	return r.sessionAccessor
}
